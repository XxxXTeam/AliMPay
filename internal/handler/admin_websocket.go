/*
Package handler 管理后台WebSocket处理器
Author: AliMPay Team
Description: 提供管理后台实时订单推送功能

功能:
  - 新订单通知
  - 订单支付通知
  - 订单过期通知
  - 统计信息推送
  - 自动广播机制

消息格式:
  {
    "type": "order_created|order_paid|order_expired|stats_update",
    "order_id": "xxx",
    "name": "商品名称",
    "payment_amount": 0.01,
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
AdminWebSocketHandler 管理后台WebSocket处理器
字段:
  - db: 数据库实例
  - upgrader: WebSocket升级器
  - connections: 连接池
  - mu: 读写锁
*/
type AdminWebSocketHandler struct {
	db          *database.DB
	upgrader    websocket.Upgrader
	connections map[*websocket.Conn]bool
	mu          sync.RWMutex
}

/*
NewAdminWebSocketHandler 创建管理后台WebSocket处理器
参数:
  - db: 数据库实例
返回:
  - *AdminWebSocketHandler: WebSocket处理器实例
*/
func NewAdminWebSocketHandler(db *database.DB) *AdminWebSocketHandler {
	handler := &AdminWebSocketHandler{
		db: db,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // 生产环境应限制来源
			},
		},
		connections: make(map[*websocket.Conn]bool),
	}

	// 订阅订单事件
	events.Subscribe(events.EventOrderCreated, func(data interface{}) {
		order, ok := data.(*model.Order)
		if ok {
			handler.broadcastOrderCreated(order)
		}
	})

	events.Subscribe(events.EventOrderPaid, func(data interface{}) {
		order, ok := data.(*model.Order)
		if ok {
			handler.broadcastOrderPaid(order)
		}
	})

	events.Subscribe(events.EventOrderExpired, func(data interface{}) {
		order, ok := data.(*model.Order)
		if ok {
			handler.broadcastOrderExpired(order)
		}
	})

	logger.Info("Admin WebSocket handler initialized with event subscriptions")

	return handler
}

/*
HandleWebSocket 处理WebSocket连接请求
参数:
  - c: Gin上下文
*/
func (h *AdminWebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade admin WebSocket connection", zap.Error(err))
		return
	}

	h.addConnection(conn)
	logger.Info("Admin WebSocket client connected", zap.String("remote_addr", conn.RemoteAddr().String()))

	// 发送初始统计信息
	go h.sendInitialStats(conn)

	// 保持连接并处理ping/pong
	go func() {
		defer func() {
			h.removeConnection(conn)
			conn.Close()
			logger.Info("Admin WebSocket client disconnected", zap.String("remote_addr", conn.RemoteAddr().String()))
		}()

		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		// 定期发送统计信息
		statsTicker := time.NewTicker(10 * time.Second)
		defer statsTicker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					logger.Error("Failed to send ping to admin client", zap.Error(err))
					return
				}
			case <-statsTicker.C:
				h.sendStats(conn)
			default:
				// 读取消息以检测断开
				_, _, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						logger.Error("Admin WebSocket read error", zap.Error(err))
					}
					return
				}
			}
		}
	}()
}

/*
sendInitialStats 发送初始统计信息
参数:
  - conn: WebSocket连接
*/
func (h *AdminWebSocketHandler) sendInitialStats(conn *websocket.Conn) {
	h.sendStats(conn)
}

/*
sendStats 发送统计信息
参数:
  - conn: WebSocket连接
*/
func (h *AdminWebSocketHandler) sendStats(conn *websocket.Conn) {
	// 查询所有订单（简化版，生产环境应该使用专门的统计方法）
	// TODO: 添加 GetOrdersByStatus 方法到数据库层
	
	// 临时实现：使用模拟数据
	message := map[string]interface{}{
		"type":          "stats_update",
		"pending_count": 0,
		"paid_count":    0,
		"total_count":   0,
		"total_amount":  0.0,
		"timestamp":     time.Now().Unix(),
	}

	h.sendMessage(conn, message)
}

/*
broadcastOrderCreated 广播订单创建事件
参数:
  - order: 订单信息
*/
func (h *AdminWebSocketHandler) broadcastOrderCreated(order *model.Order) {
	message := map[string]interface{}{
		"type":           "order_created",
		"order_id":       order.ID,
		"trade_no":       order.ID,
		"name":           order.Name,
		"payment_amount": order.PaymentAmount,
		"create_time":    order.AddTime.Format("2006-01-02 15:04:05"),
		"timestamp":      time.Now().Unix(),
	}

	h.broadcast(message)
	logger.Debug("Broadcasted order created event", zap.String("order_id", order.ID))
}

/*
broadcastOrderPaid 广播订单支付事件
参数:
  - order: 订单信息
*/
func (h *AdminWebSocketHandler) broadcastOrderPaid(order *model.Order) {
	message := map[string]interface{}{
		"type":           "order_paid",
		"order_id":       order.ID,
		"trade_no":       order.ID,
		"name":           order.Name,
		"payment_amount": order.PaymentAmount,
		"pay_time":       order.PayTime.Format("2006-01-02 15:04:05"),
		"timestamp":      time.Now().Unix(),
	}

	h.broadcast(message)
	logger.Debug("Broadcasted order paid event", zap.String("order_id", order.ID))
}

/*
broadcastOrderExpired 广播订单过期事件
参数:
  - order: 订单信息
*/
func (h *AdminWebSocketHandler) broadcastOrderExpired(order *model.Order) {
	message := map[string]interface{}{
		"type":           "order_expired",
		"order_id":       order.ID,
		"trade_no":       order.ID,
		"name":           order.Name,
		"payment_amount": order.PaymentAmount,
		"timestamp":      time.Now().Unix(),
	}

	h.broadcast(message)
	logger.Debug("Broadcasted order expired event", zap.String("order_id", order.ID))
}

/*
broadcast 广播消息给所有连接的客户端
参数:
  - message: 消息内容
*/
func (h *AdminWebSocketHandler) broadcast(message map[string]interface{}) {
	h.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(h.connections))
	for conn := range h.connections {
		connections = append(connections, conn)
	}
	h.mu.RUnlock()

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal broadcast message", zap.Error(err))
		return
	}

	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, jsonMessage); err != nil {
			logger.Error("Failed to send broadcast message", zap.Error(err))
			h.removeConnection(conn)
			conn.Close()
		}
	}
}

/*
sendMessage 发送消息给单个客户端
参数:
  - conn: WebSocket连接
  - message: 消息内容
*/
func (h *AdminWebSocketHandler) sendMessage(conn *websocket.Conn, message map[string]interface{}) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonMessage); err != nil {
		logger.Error("Failed to send message", zap.Error(err))
	}
}

/*
addConnection 添加连接
参数:
  - conn: WebSocket连接
*/
func (h *AdminWebSocketHandler) addConnection(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[conn] = true
	logger.Debug("Admin WebSocket connection added", zap.Int("total_connections", len(h.connections)))
}

/*
removeConnection 移除连接
参数:
  - conn: WebSocket连接
*/
func (h *AdminWebSocketHandler) removeConnection(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, conn)
	logger.Debug("Admin WebSocket connection removed", zap.Int("total_connections", len(h.connections)))
}

/*
GetConnectionCount 获取当前连接数
返回:
  - int: 连接数
*/
func (h *AdminWebSocketHandler) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

