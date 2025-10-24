/*
Package events 事件系统
Author: AliMPay Team
Description: 提供应用级事件发布/订阅机制

功能:
  - 解耦模块间依赖
  - 订单状态变化事件通知
  - WebSocket推送集成
  - 支持多订阅者

使用示例:

	// 订阅事件
	events.Subscribe(events.EventOrderPaid, func(data interface{}) {
	    order := data.(*model.Order)
	    // 处理订单支付事件
	})

	// 发布事件
	events.Publish(events.EventOrderPaid, order)
*/
package events

import (
	"sync"

	"alimpay-go/internal/model"
	"alimpay-go/pkg/logger"

	"go.uber.org/zap"
)

/*
事件类型定义
*/
const (
	EventOrderPaid    = "order:paid"    // 订单支付成功
	EventOrderExpired = "order:expired" // 订单过期
	EventOrderCreated = "order:created" // 订单创建
)

/*
EventHandler 事件处理函数类型
@param data 事件数据
*/
type EventHandler func(data interface{})

/*
EventBus 事件总线
功能: 管理事件订阅和发布
字段:
  - handlers: 事件处理器映射 (eventType -> []handler)
  - mu: 读写锁保护
*/
type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

/*
全局事件总线实例
*/
var globalBus = &EventBus{
	handlers: make(map[string][]EventHandler),
}

/*
Subscribe 订阅事件
功能: 注册事件处理器
参数:
  - eventType: 事件类型
  - handler: 处理函数
*/
func Subscribe(eventType string, handler EventHandler) {
	globalBus.mu.Lock()
	defer globalBus.mu.Unlock()

	globalBus.handlers[eventType] = append(globalBus.handlers[eventType], handler)

	logger.Info("Event handler subscribed",
		zap.String("event_type", eventType),
		zap.Int("total_handlers", len(globalBus.handlers[eventType])))
}

/*
Publish 发布事件
功能: 触发所有订阅该事件的处理器
参数:
  - eventType: 事件类型
  - data: 事件数据
*/
func Publish(eventType string, data interface{}) {
	globalBus.mu.RLock()
	handlers := globalBus.handlers[eventType]
	globalBus.mu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	logger.Debug("Publishing event",
		zap.String("event_type", eventType),
		zap.Int("handlers_count", len(handlers)))

	// 异步执行所有处理器
	for _, handler := range handlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Event handler panicked",
						zap.String("event_type", eventType),
						zap.Any("panic", r))
				}
			}()
			h(data)
		}(handler)
	}
}

/*
PublishOrderPaid 发布订单支付成功事件
便捷方法: 发布订单支付事件
参数:
  - order: 订单信息
*/
func PublishOrderPaid(order *model.Order) {
	Publish(EventOrderPaid, order)
}

/*
PublishOrderCreated 发布订单创建事件
便捷方法: 发布订单创建事件
参数:
  - order: 订单信息
*/
func PublishOrderCreated(order *model.Order) {
	Publish(EventOrderCreated, order)
}

/*
PublishOrderExpired 发布订单过期事件
便捷方法: 发布订单过期事件
参数:
  - order: 订单信息
*/
func PublishOrderExpired(order *model.Order) {
	Publish(EventOrderExpired, order)
}

/*
Unsubscribe 取消所有订阅
功能: 清理事件处理器（用于测试或重置）
参数:
  - eventType: 事件类型，为空则清理所有
*/
func Unsubscribe(eventType string) {
	globalBus.mu.Lock()
	defer globalBus.mu.Unlock()

	if eventType == "" {
		globalBus.handlers = make(map[string][]EventHandler)
	} else {
		delete(globalBus.handlers, eventType)
	}
}

/*
GetStats 获取事件系统统计信息
返回:
  - map[string]interface{}: 统计数据
*/
func GetStats() map[string]interface{} {
	globalBus.mu.RLock()
	defer globalBus.mu.RUnlock()

	stats := make(map[string]int)
	for eventType, handlers := range globalBus.handlers {
		stats[eventType] = len(handlers)
	}

	return map[string]interface{}{
		"subscriptions": stats,
		"event_types":   len(globalBus.handlers),
	}
}
