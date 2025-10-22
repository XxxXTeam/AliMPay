package service

import (
	"time"

	"alimpay-go/internal/database"
	"alimpay-go/internal/model"
	"alimpay-go/pkg/logger"

	"go.uber.org/zap"
)

// AutoCallbackService 自动回调服务
// 订单支付后自动触发商户回调，无需等待回调接口被调用
type AutoCallbackService struct {
	db      *database.DB
	codepay *CodePayService
	stopCh  chan struct{}
}

// NewAutoCallbackService 创建自动回调服务
func NewAutoCallbackService(db *database.DB, codepay *CodePayService) *AutoCallbackService {
	return &AutoCallbackService{
		db:      db,
		codepay: codepay,
		stopCh:  make(chan struct{}),
	}
}

// Start 启动自动回调服务
func (s *AutoCallbackService) Start() {
	go s.run()
	logger.Info("Auto callback service started")
}

// Stop 停止自动回调服务
func (s *AutoCallbackService) Stop() {
	close(s.stopCh)
	logger.Info("Auto callback service stopped")
}

// run 运行自动回调
func (s *AutoCallbackService) run() {
	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processAutoCallback()
		case <-s.stopCh:
			return
		}
	}
}

// processAutoCallback 处理自动回调
func (s *AutoCallbackService) processAutoCallback() {
	// 获取最近已支付但未回调的订单
	orders, err := s.db.GetRecentOrders(50)
	if err != nil {
		logger.Error("Failed to get recent orders", zap.Error(err))
		return
	}

	for _, order := range orders {
		// 只处理已支付的订单
		if order.Status == model.OrderStatusPaid && order.NotifyURL != "" {
			// 检查是否已发送过回调（简单检查：支付时间距现在超过10秒）
			if order.PayTime != nil && time.Since(*order.PayTime) < 10*time.Second {
				// 发送商户回调
				go func(o *model.Order) {
					logger.Info("Auto callback triggered",
						zap.String("trade_no", o.ID),
						zap.String("out_trade_no", o.OutTradeNo))

					err := s.codepay.SendNotification(o)
					if err != nil {
						logger.Error("Auto callback failed",
							zap.String("trade_no", o.ID),
							zap.Error(err))
					} else {
						logger.Info("Auto callback sent",
							zap.String("trade_no", o.ID))
					}
				}(order)
			}
		}
	}
}
