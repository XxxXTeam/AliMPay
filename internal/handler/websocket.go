/*
Package handler WebSocket处理器
Author: AliMPay Team
Description: 提供WebSocket连接管理，用于实时推送订单支付状态更新

功能:
  - 管理WebSocket连接池
  - 实时推送订单状态变化
  - 替代HTTP轮询，降低客户端压力
  - 支持多客户端订阅同一订单

连接流程:
 1. 客户端通过 /ws/order?order_id=xxx 建立连接
 2. 服务器将连接加入订阅池
 3. 订单状态更新时，推送消息给所有订阅者
 4. 连接断开时自动清理

消息格式:

	{
	  "type": "status_update",
	  "order_id": "xxx",
	  "status": 1,
	  "pay_time": "2024-01-01 12:00:00",
	  "timestamp": 1234567890
	}
*/
package handler

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"alimpay-go/internal/database"
	"alimpay-go/internal/events"
	"alimpay-go/internal/model"
	"alimpay-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

/*
WebSocketHandler WebSocket处理器
功能: 管理WebSocket连接和消息推送
字段:
  - db: 数据库实例
  - upgrader: WebSocket升级器
  - subscribers: 订单订阅者映射表 (order_id -> []*websocket.Conn)
  - mu: 读写锁，保护subscribers
*/
type WebSocketHandler struct {
	db          *database.DB
	upgrader    websocket.Upgrader
	subscribers map[string][]*websocket.Conn // order_id -> connections
	mu          sync.RWMutex
}

/*
OrderStatusMessage 订单状态消息
用途: WebSocket推送的消息格式
字段:
  - Type: 消息类型
  - OrderID: 订单号
  - Status: 订单状态 (0=待支付, 1=已支付)
  - PayTime: 支付时间
  - Timestamp: 消息时间戳
*/
type OrderStatusMessage struct {
	Type      string `json:"type"`      // 消息类型: status_update
	OrderID   string `json:"order_id"`  // 订单号
	Status    int    `json:"status"`    // 订单状态
	PayTime   string `json:"pay_time"`  // 支付时间
	Timestamp int64  `json:"timestamp"` // 时间戳
}

/*
NewWebSocketHandler 创建WebSocket处理器
参数:
  - db: 数据库实例

返回:
  - *WebSocketHandler: WebSocket处理器实例
*/
func NewWebSocketHandler(db *database.DB) *WebSocketHandler {
	handler := &WebSocketHandler{
		db: db,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// 允许跨域
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		subscribers: make(map[string][]*websocket.Conn),
	}

	// 订阅订单支付事件，自动推送给WebSocket客户端
	events.Subscribe(events.EventOrderPaid, func(data interface{}) {
		order, ok := data.(*model.Order)
		if !ok {
			return
		}
		handler.BroadcastOrderUpdate(order)
	})

	logger.Info("WebSocket handler initialized with event subscription")

	return handler
}

/*
HandleWebSocket 处理WebSocket连接请求
功能:
  - 升级HTTP连接为WebSocket
  - 订阅指定订单的状态更新
  - 维持连接并处理心跳

参数:
  - c: Gin上下文

URL参数:
  - order_id: 要订阅的订单号
*/
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	orderID := c.Query("order_id")
	if orderID == "" {
		c.JSON(400, gin.H{"error": "missing order_id parameter"})
		return
	}

	// 升级为WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade websocket",
			zap.String("order_id", orderID),
			zap.Error(err))
		return
	}

	logger.Info("WebSocket connected",
		zap.String("order_id", orderID),
		zap.String("remote_addr", c.ClientIP()))

	// 添加到订阅列表
	h.subscribe(orderID, conn)

	// 发送初始状态
	h.sendInitialStatus(conn, orderID)

	// 启动心跳和读取循环
	go h.handleConnection(conn, orderID)
}

/*
handleConnection 处理WebSocket连接的生命周期
功能:
  - 读取客户端消息(心跳)
  - 检测连接断开
  - 清理订阅

参数:
  - conn: WebSocket连接
  - orderID: 订单号
*/
func (h *WebSocketHandler) handleConnection(conn *websocket.Conn, orderID string) {
	defer func() {
		h.unsubscribe(orderID, conn)
		conn.Close()
		logger.Info("WebSocket disconnected", zap.String("order_id", orderID))
	}()

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 设置pong处理器
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 启动心跳发送
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})

	// 读取消息(主要是处理关闭)
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				close(done)
				return
			}
		}
	}()

	// 发送心跳
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

/*
sendInitialStatus 发送初始订单状态
功能: 连接建立后立即发送当前订单状态
参数:
  - conn: WebSocket连接
  - orderID: 订单号
*/
func (h *WebSocketHandler) sendInitialStatus(conn *websocket.Conn, orderID string) {
	order, err := h.db.GetOrderByID(orderID)
	if err != nil || order == nil {
		return
	}

	message := OrderStatusMessage{
		Type:      "status_update",
		OrderID:   orderID,
		Status:    order.Status,
		PayTime:   h.formatPayTime(order),
		Timestamp: time.Now().Unix(),
	}

	data, _ := json.Marshal(message)
	conn.WriteMessage(websocket.TextMessage, data)
}

/*
BroadcastOrderUpdate 广播订单状态更新
功能: 当订单状态变化时，通知所有订阅者
参数:
  - order: 更新后的订单信息
*/
func (h *WebSocketHandler) BroadcastOrderUpdate(order *model.Order) {
	h.mu.RLock()
	connections := h.subscribers[order.ID]
	h.mu.RUnlock()

	if len(connections) == 0 {
		return
	}

	message := OrderStatusMessage{
		Type:      "status_update",
		OrderID:   order.ID,
		Status:    order.Status,
		PayTime:   h.formatPayTime(order),
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	logger.Info("Broadcasting order update",
		zap.String("order_id", order.ID),
		zap.Int("subscribers", len(connections)))

	// 发送给所有订阅者
	h.mu.Lock()
	var validConns []*websocket.Conn
	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			logger.Warn("Failed to send message, removing subscriber",
				zap.Error(err))
			conn.Close()
		} else {
			validConns = append(validConns, conn)
		}
	}
	// 更新有效连接列表
	h.subscribers[order.ID] = validConns
	h.mu.Unlock()
}

/*
subscribe 订阅订单状态更新
功能: 将WebSocket连接添加到订阅列表
参数:
  - orderID: 订单号
  - conn: WebSocket连接
*/
func (h *WebSocketHandler) subscribe(orderID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.subscribers[orderID] = append(h.subscribers[orderID], conn)
	logger.Info("Subscribed to order",
		zap.String("order_id", orderID),
		zap.Int("total_subscribers", len(h.subscribers[orderID])))
}

/*
unsubscribe 取消订阅
功能: 从订阅列表中移除WebSocket连接
参数:
  - orderID: 订单号
  - conn: WebSocket连接
*/
func (h *WebSocketHandler) unsubscribe(orderID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	connections := h.subscribers[orderID]
	for i, c := range connections {
		if c == conn {
			h.subscribers[orderID] = append(connections[:i], connections[i+1:]...)
			break
		}
	}

	// 如果没有订阅者了，删除键
	if len(h.subscribers[orderID]) == 0 {
		delete(h.subscribers, orderID)
	}
}

/*
formatPayTime 格式化支付时间
参数:
  - order: 订单信息

返回:
  - string: 格式化的支付时间，如果未支付返回空字符串
*/
func (h *WebSocketHandler) formatPayTime(order *model.Order) string {
	if order.Status == model.OrderStatusPaid && !order.PayTime.IsZero() {
		return order.PayTime.Format("2006-01-02 15:04:05")
	}
	return ""
}

/*
GetStats 获取WebSocket统计信息
功能: 返回当前WebSocket连接状态
返回:
  - map[string]interface{}: 统计信息
*/
func (h *WebSocketHandler) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalConnections := 0
	for _, conns := range h.subscribers {
		totalConnections += len(conns)
	}

	return map[string]interface{}{
		"total_subscribed_orders": len(h.subscribers),
		"total_connections":       totalConnections,
	}
}
