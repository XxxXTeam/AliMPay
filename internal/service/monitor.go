package service

import (
	"fmt"
	"time"

	"github.com/alimpay/alimpay-go/internal/config"
	"github.com/alimpay/alimpay-go/internal/database"
	"github.com/alimpay/alimpay-go/internal/model"
	"github.com/alimpay/alimpay-go/pkg/lock"
	"github.com/alimpay/alimpay-go/pkg/logger"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// MonitorService 监控服务
type MonitorService struct {
	cfg              *config.Config
	db               *database.DB
	codepay          *CodePayService
	billQuery        *BillQueryService
	cron             *cron.Cron
	lockFile         string
	isRunning        bool
	apiFailureCount  int
	lastSuccessTime  time.Time
	monitoringPaused bool
}

// BillRecord 账单记录
type BillRecord struct {
	TradeNo   string
	Amount    float64
	Remark    string
	TransDate string
	Direction string
}

// NewMonitorService 创建监控服务
func NewMonitorService(cfg *config.Config, db *database.DB, codepay *CodePayService) (*MonitorService, error) {
	// 创建账单查询服务
	billQuery, err := NewBillQueryService(&cfg.Alipay)
	if err != nil {
		// 如果账单查询服务创建失败，记录警告但不中断
		logger.Warn("Failed to create bill query service, monitoring will be limited", zap.Error(err))
		billQuery = nil
	}

	return &MonitorService{
		cfg:       cfg,
		db:        db,
		codepay:   codepay,
		billQuery: billQuery,
		lockFile:  "./data/monitor.lock",
	}, nil
}

// Start 启动监控服务
func (m *MonitorService) Start() error {
	if !m.cfg.Monitor.Enabled {
		logger.Info("Monitor service is disabled")
		return nil
	}

	m.cron = cron.New()

	// 添加定时任务
	interval := m.cfg.Monitor.Interval
	spec := fmt.Sprintf("@every %ds", interval)

	_, err := m.cron.AddFunc(spec, func() {
		m.RunMonitoringCycle()
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	m.cron.Start()
	m.isRunning = true

	logger.Info("Monitor service started", zap.Int("interval", interval))
	return nil
}

// GetMonitorStatus 获取监控服务状态
func (m *MonitorService) GetMonitorStatus() map[string]interface{} {
	return map[string]interface{}{
		"running":           m.isRunning,
		"paused":            m.monitoringPaused,
		"api_failure_count": m.apiFailureCount,
		"last_success_time": m.lastSuccessTime,
		"health_status": func() string {
			if !m.isRunning {
				return "stopped"
			}
			if m.monitoringPaused {
				return "paused"
			}
			if m.apiFailureCount > 0 {
				return "degraded"
			}
			return "healthy"
		}(),
		"message": func() string {
			if m.monitoringPaused {
				return "监控已暂停（API连续失败），请使用管理后台手动处理订单"
			}
			if m.apiFailureCount > 0 {
				return fmt.Sprintf("API连续失败%d次，正在重试", m.apiFailureCount)
			}
			return "监控服务运行正常"
		}(),
	}
}

// ResumeMonitoring 恢复监控
func (m *MonitorService) ResumeMonitoring() {
	m.monitoringPaused = false
	m.apiFailureCount = 0
	logger.Info("监控服务已手动恢复")
}

// Stop 停止监控服务
func (m *MonitorService) Stop() {
	if m.cron != nil {
		m.cron.Stop()
	}
	m.isRunning = false
	logger.Info("Monitor service stopped")
}

// RunMonitoringCycle 运行一次监控周期
func (m *MonitorService) RunMonitoringCycle() {
	// 使用文件锁防止并发执行
	fileLock := lock.NewFileLock(m.lockFile, time.Duration(m.cfg.Monitor.LockTimeout)*time.Second)

	acquired, err := fileLock.TryLock()
	if err != nil {
		logger.Error("Failed to acquire lock", zap.Error(err))
		return
	}

	if !acquired {
		logger.Debug("Another monitoring cycle is running, skipping")
		return
	}
	defer fileLock.Unlock()

	logger.Info("Starting payment monitoring cycle")

	// 1. 清理过期订单
	count, err := m.codepay.CleanupExpiredOrders()
	if err != nil {
		logger.Error("Failed to cleanup expired orders", zap.Error(err))
	} else if count > 0 {
		logger.Info("Cleaned up expired orders", zap.Int64("count", count))
	}

	// 2. 查询最近的账单
	// 注意：这里简化了支付宝账单查询部分
	// 在实际应用中，需要集成支付宝OpenAPI SDK来查询账单
	bills := m.queryRecentBills()

	if len(bills) == 0 {
		logger.Debug("No recent bills found")
		return
	}

	logger.Info("Found bills to process", zap.Int("count", len(bills)))

	// 3. 处理账单
	if m.cfg.Payment.BusinessQRMode.Enabled {
		m.processBillsForBusinessMode(bills)
	} else {
		m.processBillsForTraditionalMode(bills)
	}

	logger.Info("Payment monitoring cycle completed")
}

// queryRecentBills 查询最近的账单
func (m *MonitorService) queryRecentBills() []BillRecord {
	if m.billQuery == nil {
		logger.Debug("Bill query service not available")
		return []BillRecord{}
	}

	// 查询配置的时间范围
	minutes := m.cfg.Payment.QueryMinutesBack
	if minutes == 0 {
		minutes = 30
	}

	// 转换为小时（向上取整）
	hours := (minutes + 59) / 60

	// 查询账单
	result, err := m.billQuery.QueryRecentBills(hours)
	if err != nil {
		m.apiFailureCount++
		logger.Error("Failed to query bills",
			zap.Error(err),
			zap.Int("failure_count", m.apiFailureCount))

		// 连续失败5次后暂停监控，避免频繁失败
		if m.apiFailureCount >= 5 && !m.monitoringPaused {
			m.monitoringPaused = true
			logger.Warn("监控服务已暂停（API连续失败5次）",
				zap.String("建议", "请检查支付宝API配置或使用管理后台手动处理订单"))
			logger.Warn("管理后台地址", zap.String("url", "http://localhost:8080/admin/dashboard"))
		}

		return []BillRecord{}
	}

	// 查询成功，重置失败计数
	if m.apiFailureCount > 0 || m.monitoringPaused {
		logger.Info("支付宝API恢复正常",
			zap.Int("previous_failures", m.apiFailureCount))
		m.apiFailureCount = 0
		m.monitoringPaused = false
	}
	m.lastSuccessTime = time.Now()

	// 提取账单数据
	success, _ := result["success"].(bool)
	if !success {
		logger.Warn("Bill query was not successful")
		return []BillRecord{}
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		logger.Warn("Invalid bill data structure")
		return []BillRecord{}
	}

	detailList, ok := data["detail_list"].([]map[string]interface{})
	if !ok {
		logger.Warn("Invalid detail_list structure")
		return []BillRecord{}
	}

	// 转换为 BillRecord 格式
	var bills []BillRecord
	for _, detail := range detailList {
		// 只处理收入类型
		direction, _ := detail["direction"].(string)
		if direction != "收入" {
			continue
		}

		amountStr, _ := detail["trans_amount"].(string)
		var amount float64
		fmt.Sscanf(amountStr, "%f", &amount)

		bill := BillRecord{
			TradeNo:   detail["alipay_order_no"].(string),
			Amount:    amount,
			Remark:    detail["trans_memo"].(string),
			TransDate: detail["trans_dt"].(string),
			Direction: direction,
		}
		bills = append(bills, bill)
	}

	logger.Info("Query recent bills completed", zap.Int("count", len(bills)))
	return bills
}

// processBillsForBusinessMode 处理经营码模式的账单
func (m *MonitorService) processBillsForBusinessMode(bills []BillRecord) {
	logger.Info("Processing bills for business QR mode")

	for _, bill := range bills {
		// 只处理收入类型的账单
		if bill.Direction != "收入" {
			continue
		}

		logger.Info("Processing bill",
			zap.String("trade_no", bill.TradeNo),
			zap.Float64("amount", bill.Amount),
			zap.String("trans_date", bill.TransDate))

		// 根据金额查找待支付订单
		order, err := m.db.GetPendingOrderByAmount(bill.Amount)
		if err != nil {
			logger.Error("Failed to get pending order", zap.Error(err))
			continue
		}

		if order == nil {
			logger.Debug("No pending order found for amount", zap.Float64("amount", bill.Amount))
			continue
		}

		// 验证时间容差
		tolerance := time.Duration(m.cfg.Payment.BusinessQRMode.MatchTolerance) * time.Second
		billTime, _ := time.Parse("2006-01-02 15:04:05", bill.TransDate)
		timeDiff := billTime.Sub(order.AddTime)

		if timeDiff < 0 || timeDiff > tolerance {
			logger.Warn("Order found but outside time tolerance",
				zap.String("order_id", order.ID),
				zap.Duration("time_diff", timeDiff),
				zap.Duration("tolerance", tolerance))
			continue
		}

		// 更新订单状态
		if err := m.updateOrderToPaid(order); err != nil {
			logger.Error("Failed to update order status", zap.Error(err))
			continue
		}

		logger.Info("Order paid successfully",
			zap.String("order_id", order.ID),
			zap.String("out_trade_no", order.OutTradeNo),
			zap.Float64("amount", bill.Amount))
	}
}

// processBillsForTraditionalMode 处理传统模式的账单
func (m *MonitorService) processBillsForTraditionalMode(bills []BillRecord) {
	logger.Info("Processing bills for traditional mode")

	for _, bill := range bills {
		// 只处理收入类型的账单
		if bill.Direction != "收入" {
			continue
		}

		if bill.Remark == "" {
			logger.Debug("Skipping bill with empty remark", zap.String("trade_no", bill.TradeNo))
			continue
		}

		logger.Info("Processing bill",
			zap.String("trade_no", bill.TradeNo),
			zap.Float64("amount", bill.Amount),
			zap.String("remark", bill.Remark))

		// 根据备注（订单号）查找订单
		outTradeNo := bill.Remark
		order, err := m.db.GetOrderByOutTradeNo(outTradeNo, m.codepay.GetMerchantID())
		if err != nil {
			logger.Error("Failed to get order", zap.Error(err))
			continue
		}

		if order == nil {
			logger.Debug("No order found for remark", zap.String("remark", bill.Remark))
			continue
		}

		if order.Status == model.OrderStatusPaid {
			logger.Debug("Order already paid", zap.String("order_id", order.ID))
			continue
		}

		// 验证金额
		if fmt.Sprintf("%.2f", order.Price) != fmt.Sprintf("%.2f", bill.Amount) {
			logger.Warn("Amount mismatch",
				zap.String("order_id", order.ID),
				zap.Float64("expected", order.Price),
				zap.Float64("actual", bill.Amount))
			continue
		}

		// 更新订单状态
		if err := m.updateOrderToPaid(order); err != nil {
			logger.Error("Failed to update order status", zap.Error(err))
			continue
		}

		logger.Info("Order paid successfully",
			zap.String("order_id", order.ID),
			zap.String("out_trade_no", order.OutTradeNo),
			zap.Float64("amount", bill.Amount))
	}
}

// updateOrderToPaid 更新订单为已支付状态
func (m *MonitorService) updateOrderToPaid(order *model.Order) error {
	// 更新订单状态
	if err := m.db.UpdateOrderStatus(order.ID, model.OrderStatusPaid, time.Now()); err != nil {
		return err
	}

	// 发送通知给商户
	if err := m.codepay.SendNotification(order); err != nil {
		logger.Warn("Failed to send notification", zap.Error(err))
	}

	return nil
}

// GetStatus 获取监控服务状态
func (m *MonitorService) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":   m.cfg.Monitor.Enabled,
		"running":   m.isRunning,
		"interval":  m.cfg.Monitor.Interval,
		"lock_file": m.lockFile,
	}
}
