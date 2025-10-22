package service

import (
	"fmt"
	"time"

	"alimpay-go/internal/config"
	"alimpay-go/internal/database"
	"alimpay-go/internal/model"
	"alimpay-go/pkg/lock"
	"alimpay-go/pkg/logger"

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

	logger.Info("========== 开始支付监听周期 ==========")

	// 1. 清理过期订单
	count, err := m.codepay.CleanupExpiredOrders()
	if err != nil {
		logger.Error("Failed to cleanup expired orders", zap.Error(err))
	} else if count > 0 {
		logger.Info("✓ 清理过期订单", zap.Int64("count", count))
	}

	// 2. 查询10分钟内创建的待支付订单（只监听10分钟）
	pendingOrders, err := m.getRecentPendingOrders(10 * time.Minute)
	if err != nil {
		logger.Error("Failed to get pending orders", zap.Error(err))
		return
	}

	logger.Info("📋 待支付订单统计（10分钟内创建）",
		zap.Int("total_count", len(pendingOrders)))

	if len(pendingOrders) == 0 {
		logger.Info("✅ 没有需要监听的订单，跳过本次查询（仅监听10分钟内创建的订单）")
		return
	}

	// 输出每个待支付订单的详细信息
	for _, order := range pendingOrders {
		orderAge := time.Since(order.AddTime)
		remainingTime := 10*time.Minute - orderAge

		logger.Info("🔍 待支付订单详情",
			zap.String("订单号", order.ID),
			zap.String("商户订单号", order.OutTradeNo),
			zap.Float64("应付金额", order.Price),
			zap.Float64("实付金额", order.PaymentAmount),
			zap.String("创建时间", order.AddTime.Format("2006-01-02 15:04:05")),
			zap.Duration("已等待", orderAge),
			zap.Duration("剩余监听时间", remainingTime),
			zap.String("状态", "监听中"))
	}

	// 3. 查询最近的账单
	// 注意：这里简化了支付宝账单查询部分
	// 在实际应用中，需要集成支付宝OpenAPI SDK来查询账单
	bills := m.queryRecentBills()

	if len(bills) == 0 {
		logger.Info("⚠️  支付宝账单查询无结果（可能是API未上线），请使用管理后台手动标记")
		logger.Info("管理后台地址: http://localhost:8080/admin/dashboard")
		return
	}

	logger.Info("💰 查询到支付宝账单", zap.Int("count", len(bills)))

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

// getPendingOrders 获取待支付订单
func (m *MonitorService) getPendingOrders() ([]*model.Order, error) {
	// 获取最近10分钟的待支付订单
	since := time.Now().Add(-10 * time.Minute)
	orders, err := m.db.GetPendingOrdersSince(since)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}

	return orders, nil
}

// getRecentPendingOrders 获取指定时间范围内的待支付订单
func (m *MonitorService) getRecentPendingOrders(duration time.Duration) ([]*model.Order, error) {
	since := time.Now().Add(-duration)
	orders, err := m.db.GetPendingOrdersSince(since)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}

	// 过滤掉超过监听时间的订单
	var recentOrders []*model.Order
	now := time.Now()
	for _, order := range orders {
		orderAge := now.Sub(order.AddTime)
		if orderAge <= duration {
			recentOrders = append(recentOrders, order)
		} else {
			logger.Info("⏰ 订单超过监听时间，不再监听",
				zap.String("订单号", order.ID),
				zap.String("商户订单号", order.OutTradeNo),
				zap.Duration("订单年龄", orderAge),
				zap.Duration("监听时长", duration),
				zap.String("说明", "订单创建超过10分钟，将由管理员手动处理"))
		}
	}

	return recentOrders, nil
}

// processBillsForBusinessMode 处理经营码模式的账单
func (m *MonitorService) processBillsForBusinessMode(bills []BillRecord) {
	logger.Info("🏪 处理经营码模式账单", zap.Int("总账单数", len(bills)))

	for _, bill := range bills {
		// 只处理收入类型的账单
		if bill.Direction != "收入" {
			continue
		}

		logger.Info("💳 匹配支付记录",
			zap.String("支付宝订单号", bill.TradeNo),
			zap.Float64("金额", bill.Amount),
			zap.String("交易时间", bill.TransDate),
			zap.String("备注", bill.Remark))

		// 根据金额查找待支付订单
		order, err := m.db.GetPendingOrderByAmount(bill.Amount)
		if err != nil {
			logger.Error("Failed to get pending order", zap.Error(err))
			continue
		}

		if order == nil {
			logger.Info("❌ 未找到匹配订单", zap.Float64("金额", bill.Amount))
			continue
		}

		logger.Info("✓ 找到匹配订单",
			zap.String("订单号", order.ID),
			zap.String("商户订单号", order.OutTradeNo),
			zap.Float64("订单金额", order.PaymentAmount),
			zap.String("订单创建时间", order.AddTime.Format("2006-01-02 15:04:05")))

		// 验证时间容差
		tolerance := time.Duration(m.cfg.Payment.BusinessQRMode.MatchTolerance) * time.Second

		// 解析支付时间（使用北京时间）
		billTime, err := time.ParseInLocation("2006-01-02 15:04:05", bill.TransDate, time.Local)
		if err != nil {
			logger.Error("❌ 解析支付时间失败",
				zap.String("支付时间", bill.TransDate),
				zap.Error(err))
			continue
		}

		timeDiff := billTime.Sub(order.AddTime)

		logger.Info("⏰ 时间对比",
			zap.String("支付时间", bill.TransDate),
			zap.String("订单时间", order.AddTime.Format("2006-01-02 15:04:05")),
			zap.Duration("时间差", timeDiff))

		// 支付时间必须晚于订单创建时间（正常流程：先创建订单，后支付）
		if timeDiff < 0 {
			logger.Warn("❌ 支付时间早于订单创建时间，跳过",
				zap.String("订单号", order.ID),
				zap.String("支付时间", bill.TransDate),
				zap.String("订单创建时间", order.AddTime.Format("2006-01-02 15:04:05")),
				zap.Duration("时间差", timeDiff),
				zap.String("说明", "这笔支付可能是其他订单的，不匹配当前订单"))
			continue
		}

		// 时间差不能超过容差
		if timeDiff > tolerance {
			logger.Warn("⏰ 订单超时容差",
				zap.String("订单号", order.ID),
				zap.Duration("时间差", timeDiff),
				zap.Duration("容差", tolerance),
				zap.String("说明", "支付时间与订单创建时间差距过大"))
			continue
		}

		logger.Info("✓ 时间验证通过", zap.Duration("延迟时间", timeDiff))

		// 更新订单状态
		if err := m.updateOrderToPaid(order); err != nil {
			logger.Error("❌ 更新订单状态失败", zap.Error(err))
			continue
		}

		logger.Info("✅ 订单支付成功",
			zap.String("订单号", order.ID),
			zap.String("商户订单号", order.OutTradeNo),
			zap.Float64("金额", bill.Amount),
			zap.String("支付宝订单号", bill.TradeNo))
	}
}

// processBillsForTraditionalMode 处理传统模式的账单
func (m *MonitorService) processBillsForTraditionalMode(bills []BillRecord) {
	logger.Info("🔐 处理传统模式账单", zap.Int("总账单数", len(bills)))

	for _, bill := range bills {
		// 只处理收入类型的账单
		if bill.Direction != "收入" {
			continue
		}

		if bill.Remark == "" {
			logger.Debug("跳过无备注账单", zap.String("支付宝订单号", bill.TradeNo))
			continue
		}

		logger.Info("💳 处理支付记录",
			zap.String("支付宝订单号", bill.TradeNo),
			zap.Float64("金额", bill.Amount),
			zap.String("备注/订单号", bill.Remark))

		// 根据备注（订单号）查找订单
		outTradeNo := bill.Remark
		order, err := m.db.GetOrderByOutTradeNo(outTradeNo, m.codepay.GetMerchantID())
		if err != nil {
			logger.Error("Failed to get order", zap.Error(err))
			continue
		}

		if order == nil {
			logger.Info("❌ 未找到订单", zap.String("商户订单号", bill.Remark))
			continue
		}

		logger.Info("✓ 找到订单",
			zap.String("订单号", order.ID),
			zap.String("商户订单号", order.OutTradeNo),
			zap.Int("状态", order.Status))

		if order.Status == model.OrderStatusPaid {
			logger.Info("ℹ️  订单已支付，跳过", zap.String("订单号", order.ID))
			continue
		}

		// 验证金额
		if fmt.Sprintf("%.2f", order.Price) != fmt.Sprintf("%.2f", bill.Amount) {
			logger.Warn("💰 金额不匹配",
				zap.String("订单号", order.ID),
				zap.Float64("期望金额", order.Price),
				zap.Float64("实际金额", bill.Amount))
			continue
		}

		logger.Info("✓ 金额验证通过", zap.Float64("金额", bill.Amount))

		// 更新订单状态
		if err := m.updateOrderToPaid(order); err != nil {
			logger.Error("❌ 更新订单状态失败", zap.Error(err))
			continue
		}

		logger.Info("✅ 订单支付成功",
			zap.String("订单号", order.ID),
			zap.String("商户订单号", order.OutTradeNo),
			zap.Float64("金额", bill.Amount),
			zap.String("支付宝订单号", bill.TradeNo))
	}
}

// updateOrderToPaid 更新订单为已支付状态
func (m *MonitorService) updateOrderToPaid(order *model.Order) error {
	logger.Info("🔄 更新订单状态为已支付",
		zap.String("订单号", order.ID),
		zap.String("商户订单号", order.OutTradeNo))

	// 更新订单状态
	payTime := time.Now()
	if err := m.db.UpdateOrderStatus(order.ID, model.OrderStatusPaid, payTime); err != nil {
		return err
	}

	logger.Info("✓ 订单状态已更新",
		zap.String("订单号", order.ID),
		zap.String("支付时间", payTime.Format("2006-01-02 15:04:05")))

	// 发送通知给商户
	logger.Info("📤 发送支付通知给商户",
		zap.String("订单号", order.ID),
		zap.String("通知URL", order.NotifyURL))

	if err := m.codepay.SendNotification(order); err != nil {
		logger.Warn("⚠️  发送通知失败（将在后台自动重试）", zap.Error(err))
	} else {
		logger.Info("✓ 通知发送成功", zap.String("订单号", order.ID))
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
