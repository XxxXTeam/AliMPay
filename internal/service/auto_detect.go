package service

import (
	"fmt"
	"sync"
	"time"

	"alimpay-go/internal/database"
	"alimpay-go/internal/model"
	"alimpay-go/internal/pkg/logger"

	"go.uber.org/zap"
)

// AutoDetectService 自动检测服务（不依赖支付宝API的备用方案）
// 原理：通过数据库轮询检查订单状态，配合前端实时查询和管理后台确认
type AutoDetectService struct {
	db                 *database.DB
	codepay            *CodePayService
	running            bool
	stopChan           chan struct{}
	mu                 sync.Mutex
	checkInterval      time.Duration
	orderCheckDuration time.Duration // 订单检查时长（超过此时间的订单不再检查）
}

// NewAutoDetectService 创建自动检测服务
func NewAutoDetectService(db *database.DB, codepay *CodePayService) *AutoDetectService {
	return &AutoDetectService{
		db:                 db,
		codepay:            codepay,
		stopChan:           make(chan struct{}),
		checkInterval:      5 * time.Second,  // 每5秒检查一次
		orderCheckDuration: 10 * time.Minute, // 检查10分钟内的订单
	}
}

// Start 启动自动检测服务
func (s *AutoDetectService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	logger.Info("Auto detect service started")

	go s.run()
}

// Stop 停止自动检测服务
func (s *AutoDetectService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	logger.Info("Auto detect service stopped")
}

// run 运行检测循环
func (s *AutoDetectService) run() {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkPendingOrders()
		case <-s.stopChan:
			return
		}
	}
}

// checkPendingOrders 检查待支付订单
func (s *AutoDetectService) checkPendingOrders() {
	// 获取最近N分钟内的待支付订单
	since := time.Now().Add(-s.orderCheckDuration)
	orders, err := s.db.GetPendingOrdersSince(since)
	if err != nil {
		logger.Error("Failed to get pending orders", zap.Error(err))
		return
	}

	if len(orders) == 0 {
		return
	}

	logger.Debug("Checking pending orders", zap.Int("count", len(orders)))

	for _, order := range orders {
		// 检查订单是否超时
		orderAge := time.Since(order.AddTime)
		if orderAge > time.Duration(s.codepay.cfg.Payment.OrderTimeout)*time.Second {
			logger.Info("Order timeout, will be cleaned up",
				zap.String("order_id", order.ID),
				zap.String("out_trade_no", order.OutTradeNo),
				zap.Duration("age", orderAge))
			continue
		}

		// 这里可以实现其他检测逻辑
		// 例如：检查是否有手动标记、检查缓存状态等
		s.checkOrderViaCache(order)
	}
}

// checkOrderViaCache 通过缓存检查订单（可扩展）
func (s *AutoDetectService) checkOrderViaCache(order *model.Order) {
	// TODO: 实现基于缓存的订单状态检查
	// 例如：Redis中存储的支付状态
	// 例如：消息队列中的支付通知
}

// MarkOrderPaidManually 手动标记订单已支付（供管理后台调用）
func (s *AutoDetectService) MarkOrderPaidManually(outTradeNo, pid string) error {
	order, err := s.db.GetOrderByOutTradeNo(outTradeNo, pid)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return fmt.Errorf("order not found")
	}

	if order.Status == model.OrderStatusPaid {
		return fmt.Errorf("order already paid")
	}

	// 更新订单状态
	payTime := time.Now()
	if err := s.db.UpdateOrderStatus(order.ID, model.OrderStatusPaid, payTime); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	logger.Info("Order marked as paid manually",
		zap.String("order_id", order.ID),
		zap.String("out_trade_no", order.OutTradeNo))

	// 发送通知
	go func() {
		if err := s.codepay.SendNotification(order); err != nil {
			logger.Error("Failed to send notification",
				zap.String("order_id", order.ID),
				zap.Error(err))
		}
	}()

	return nil
}

// GetPendingOrdersCount 获取待支付订单数量
func (s *AutoDetectService) GetPendingOrdersCount() (int, error) {
	since := time.Now().Add(-s.orderCheckDuration)
	orders, err := s.db.GetPendingOrdersSince(since)
	if err != nil {
		return 0, err
	}
	return len(orders), nil
}

// GetStatus 获取服务状态
func (s *AutoDetectService) GetStatus() map[string]interface{} {
	s.mu.Lock()
	running := s.running
	s.mu.Unlock()

	pendingCount, _ := s.GetPendingOrdersCount()

	return map[string]interface{}{
		"running":              running,
		"check_interval":       s.checkInterval.Seconds(),
		"order_check_duration": s.orderCheckDuration.Minutes(),
		"pending_orders_count": pendingCount,
		"description":          "自动检测服务（备用方案）",
		"features": []string{
			"实时检测待支付订单",
			"自动清理超时订单",
			"支持手动标记支付",
			"配合管理后台使用",
		},
	}
}
