// Package service 订单监听服务
// @author AliMPay Team
// @description 提供订单支付状态监听功能，使用Worker池提高性能
package service

import (
	"fmt"
	"time"

	"alimpay-go/internal/config"
	"alimpay-go/internal/database"
	"alimpay-go/internal/events"
	"alimpay-go/internal/model"
	"alimpay-go/internal/worker"
	"alimpay-go/pkg/lock"
	"alimpay-go/pkg/logger"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// BillRecord 账单记录
// @description 支付宝账单数据结构
type BillRecord struct {
	TradeNo   string  // 支付宝订单号
	Amount    float64 // 金额
	Remark    string  // 备注
	TransDate string  // 交易时间
	Direction string  // 方向（收入/支出）
}

// MonitorService 订单监听服务
// @description 定期检查待支付订单，使用Worker池处理订单监听任务
type MonitorService struct {
	cfg              *config.Config
	db               *database.DB
	codepay          *CodePayService
	billQuery        *BillQueryService            // 默认账单查询服务（使用全局配置）
	qrBillQueries    map[string]*BillQueryService // 二维码专属的账单查询服务 (qr_id -> service)
	workerPool       *worker.Pool
	cron             *cron.Cron
	lockFile         string
	isRunning        bool
	apiFailureCount  int
	lastSuccessTime  time.Time
	monitoringPaused bool
}

// NewMonitorService 创建监听服务
// @description 初始化订单监听服务和Worker池
// @param cfg 配置
// @param db 数据库实例
// @param codepay 码支付服务
// @return *MonitorService 监听服务实例
// @return error 创建错误
func NewMonitorService(cfg *config.Config, db *database.DB, codepay *CodePayService) (*MonitorService, error) {
	// 创建默认账单查询服务（使用全局配置）
	billQuery, err := NewBillQueryService(&cfg.Alipay)
	if err != nil {
		logger.Warn("Failed to create bill query service, monitoring will be limited", zap.Error(err))
		billQuery = nil
	}

	// 为配置了独立API的二维码创建专属的账单查询服务
	qrBillQueries := make(map[string]*BillQueryService)
	if cfg.Payment.BusinessQRMode.Enabled && len(cfg.Payment.BusinessQRMode.QRCodePaths) > 0 {
		for _, qrCode := range cfg.Payment.BusinessQRMode.QRCodePaths {
			if qrCode.HasIndependentAPI() {
				// 获取该二维码的有效配置
				qrAlipayConfig := qrCode.GetEffectiveAlipayConfig(&cfg.Alipay)

				// 创建专属的账单查询服务
				qrBillQuery, err := NewBillQueryService(qrAlipayConfig)
				if err != nil {
					logger.Warn("Failed to create bill query service for QR code",
						zap.String("qr_id", qrCode.ID),
						zap.Error(err))
					continue
				}

				qrBillQueries[qrCode.ID] = qrBillQuery
				logger.Info("Created independent bill query service for QR code",
					zap.String("qr_id", qrCode.ID),
					zap.String("app_id", qrAlipayConfig.AppID))
			}
		}
	}

	// 创建Worker池 - 使用固定数量的Worker避免创建过多goroutine
	// workerCount: 5个Worker足够处理大部分场景
	// queueSize: 队列大小为100，可容纳100个待处理订单
	workerPool := worker.NewPool(5, 100)

	return &MonitorService{
		cfg:           cfg,
		db:            db,
		codepay:       codepay,
		billQuery:     billQuery,
		qrBillQueries: qrBillQueries,
		workerPool:    workerPool,
		lockFile:      "./data/monitor.lock",
	}, nil
}

// Start 启动监听服务
// @description 启动定时任务和Worker池
// @return error 启动错误
func (m *MonitorService) Start() error {
	if !m.cfg.Monitor.Enabled {
		logger.Info("Monitor service is disabled")
		return nil
	}

	// 启动Worker池
	m.workerPool.Start()

	// 创建定时任务
	m.cron = cron.New()

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

	logger.Success("Monitor service started",
		zap.Int("interval_seconds", interval),
		zap.String("monitor_mode", func() string {
			if m.cfg.Payment.BusinessQRMode.Enabled {
				return "business_qr"
			}
			return "traditional"
		}()))

	return nil
}

// Stop 停止监听服务
// @description 停止定时任务和Worker池
func (m *MonitorService) Stop() {
	if m.cron != nil {
		m.cron.Stop()
	}

	if m.workerPool != nil {
		m.workerPool.Stop()
	}

	m.isRunning = false
	logger.Info("Monitor service stopped")
}

// RunMonitoringCycle 运行一次监听周期
// @description 获取待支付订单并提交到Worker池处理
func (m *MonitorService) RunMonitoringCycle() {
	// 使用文件锁防止并发执行
	fileLock := lock.NewFileLock(m.lockFile, time.Duration(m.cfg.Monitor.LockTimeout)*time.Second)

	acquired, err := fileLock.TryLock()
	if err != nil {
		logger.Error("Failed to acquire lock", zap.Error(err))
		return
	}

	if !acquired {
		return // 另一个周期正在运行
	}
	defer func() {
		if err := fileLock.Unlock(); err != nil {
			logger.Error("Failed to unlock file", zap.Error(err))
		}
	}()

	// 1. 清理过期订单
	if m.cfg.Payment.AutoCleanup {
		count, err := m.codepay.CleanupExpiredOrders()
		if err != nil {
			logger.Error("Failed to cleanup expired orders", zap.Error(err))
		} else if count > 0 {
			logger.Info("Cleaned up expired orders", zap.Int64("count", count))
		}
	}

	// 2. 获取待支付订单（只监听10分钟内创建的订单）
	pendingOrders, err := m.getRecentPendingOrders(10 * time.Minute)
	if err != nil {
		logger.Error("Failed to get pending orders", zap.Error(err))
		return
	}

	if len(pendingOrders) == 0 {
		return // 没有待支付订单
	}

	logger.Info("Found pending orders to monitor",
		zap.Int("count", len(pendingOrders)))

	// 3. 提交订单到Worker池处理
	submitted := 0
	rejected := 0

	for _, order := range pendingOrders {
		task := NewOrderMonitorTask(order, m)

		err := m.workerPool.Submit(task)
		if err != nil {
			rejected++
			if err == worker.ErrQueueFull {
				logger.Warn("Worker pool queue full, task rejected",
					zap.String("order_id", order.ID))
			}
		} else {
			submitted++
		}
	}

	if submitted > 0 {
		logger.Info("Submitted orders to worker pool",
			zap.Int("submitted", submitted),
			zap.Int("rejected", rejected))
	}
}

// GetBillQueryServiceForOrder 获取订单对应的账单查询服务
// @description 根据订单的二维码ID返回对应的账单查询服务
// @param order 订单
// @return *BillQueryService 账单查询服务
func (m *MonitorService) GetBillQueryServiceForOrder(order *model.Order) *BillQueryService {
	// 如果订单有分配的二维码ID，尝试使用对应的专属服务
	if order.QRCodeID != "" {
		if qrBillQuery, exists := m.qrBillQueries[order.QRCodeID]; exists {
			logger.Debug("Using QR code specific bill query service",
				zap.String("order_id", order.ID),
				zap.String("qr_code_id", order.QRCodeID))
			return qrBillQuery
		}
	}

	// 否则使用默认服务
	return m.billQuery
}

// queryRecentBills 查询最近的账单（使用默认服务）
// @description 从支付宝查询最近的收入账单
// @return []BillRecord 账单列表
// @return error 查询错误
func (m *MonitorService) queryRecentBills() ([]BillRecord, error) {
	if m.billQuery == nil {
		return []BillRecord{}, nil
	}

	// 查询最近1小时的账单
	result, err := m.billQuery.QueryRecentBills(1)
	if err != nil {
		m.apiFailureCount++
		logger.Error("Failed to query bills",
			zap.Error(err),
			zap.Int("failure_count", m.apiFailureCount))

		if m.apiFailureCount >= 5 && !m.monitoringPaused {
			m.monitoringPaused = true
			logger.Warn("Monitoring paused due to API failures",
				zap.Int("failures", m.apiFailureCount))
		}

		return []BillRecord{}, err
	}

	// 查询成功，重置失败计数
	if m.apiFailureCount > 0 || m.monitoringPaused {
		logger.Info("Alipay API recovered", zap.Int("previous_failures", m.apiFailureCount))
		m.apiFailureCount = 0
		m.monitoringPaused = false
	}
	m.lastSuccessTime = time.Now()

	// 解析账单数据
	success, _ := result["success"].(bool)
	if !success {
		return []BillRecord{}, nil
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return []BillRecord{}, nil
	}

	detailList, ok := data["detail_list"].([]map[string]interface{})
	if !ok {
		return []BillRecord{}, nil
	}

	var bills []BillRecord
	for _, detail := range detailList {
		direction, _ := detail["direction"].(string)
		if direction != "收入" {
			continue
		}

		amountStr, _ := detail["trans_amount"].(string)
		var amount float64
		if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
			logger.Warn("Failed to parse amount",
				zap.String("amount_str", amountStr),
				zap.Error(err))
			continue
		}

		bill := BillRecord{
			TradeNo:   detail["alipay_order_no"].(string),
			Amount:    amount,
			Remark:    detail["trans_memo"].(string),
			TransDate: detail["trans_dt"].(string),
			Direction: direction,
		}
		bills = append(bills, bill)
	}

	return bills, nil
}

// queryRecentBillsForQRCode 查询特定二维码的最近账单
// @description 使用二维码专属的API查询账单
// @param qrCodeID 二维码ID
// @return []BillRecord 账单列表
// @return error 查询错误
func (m *MonitorService) queryRecentBillsForQRCode(qrCodeID string) ([]BillRecord, error) {
	// 获取二维码专属的账单查询服务
	qrBillQuery, exists := m.qrBillQueries[qrCodeID]
	if !exists {
		// 如果没有专属服务，使用默认服务
		return m.queryRecentBills()
	}

	// 查询最近1小时的账单
	result, err := qrBillQuery.QueryRecentBills(1)
	if err != nil {
		logger.Error("Failed to query bills for QR code",
			zap.String("qr_code_id", qrCodeID),
			zap.Error(err))
		return []BillRecord{}, err
	}

	// 解析账单数据
	success, _ := result["success"].(bool)
	if !success {
		return []BillRecord{}, nil
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return []BillRecord{}, nil
	}

	detailList, ok := data["detail_list"].([]map[string]interface{})
	if !ok {
		return []BillRecord{}, nil
	}

	var bills []BillRecord
	for _, detail := range detailList {
		direction, _ := detail["direction"].(string)
		if direction != "收入" {
			continue
		}

		amountStr, _ := detail["trans_amount"].(string)
		var amount float64
		if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
			logger.Warn("Failed to parse amount for QR code",
				zap.String("qr_code_id", qrCodeID),
				zap.String("amount_str", amountStr),
				zap.Error(err))
			continue
		}

		bill := BillRecord{
			TradeNo:   detail["alipay_order_no"].(string),
			Amount:    amount,
			Remark:    detail["trans_memo"].(string),
			TransDate: detail["trans_dt"].(string),
			Direction: direction,
		}
		bills = append(bills, bill)
	}

	logger.Debug("Queried bills for QR code",
		zap.String("qr_code_id", qrCodeID),
		zap.Int("bill_count", len(bills)))

	return bills, nil
}

// updateOrderToPaid 更新订单为已支付状态
// @description 更新数据库并发送商户通知
// @param order 订单
// @param alipayTradeNo 支付宝订单号
// @return error 更新错误
func (m *MonitorService) updateOrderToPaid(order *model.Order, alipayTradeNo string) error {
	payTime := time.Now()

	if err := m.db.UpdateOrderStatus(order.ID, model.OrderStatusPaid, payTime); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	logger.Success("Order paid successfully",
		zap.String("order_id", order.ID),
		zap.String("merchant_order_no", order.OutTradeNo),
		zap.Float64("amount", order.PaymentAmount),
		zap.String("alipay_trade_no", alipayTradeNo))

	// 重新获取更新后的订单信息
	updatedOrder, err := m.db.GetOrderByID(order.ID)
	if err == nil && updatedOrder != nil {
		// 发布订单支付事件（触发WebSocket推送等）
		events.PublishOrderPaid(updatedOrder)
	}

	// 发送通知给商户
	if err := m.codepay.SendNotification(order); err != nil {
		logger.Warn("Failed to send notification (will retry later)",
			zap.String("order_id", order.ID),
			zap.Error(err))
	}

	return nil
}

// getRecentPendingOrders 获取最近的待支付订单
// @description 查询指定时间范围内创建的待支付订单
// @param duration 时间范围
// @return []*model.Order 订单列表
// @return error 查询错误
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
		}
	}

	return recentOrders, nil
}

// GetMonitorStatus 获取监听服务状态
// @description 返回监听服务的当前运行状态
// @return map[string]interface{} 状态信息
func (m *MonitorService) GetMonitorStatus() map[string]interface{} {
	stats := m.workerPool.GetStats()

	return map[string]interface{}{
		"running":           m.isRunning,
		"paused":            m.monitoringPaused,
		"api_failure_count": m.apiFailureCount,
		"last_success_time": m.lastSuccessTime,
		"worker_pool":       stats,
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

// ResumeMonitoring 恢复监听
// @description 手动恢复被暂停的监听服务
func (m *MonitorService) ResumeMonitoring() {
	m.monitoringPaused = false
	m.apiFailureCount = 0
	logger.Info("Monitoring service resumed manually")
}

// GetStatus 获取服务状态
// @description 返回服务的基本配置信息
// @return map[string]interface{} 状态信息
func (m *MonitorService) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":   m.cfg.Monitor.Enabled,
		"running":   m.isRunning,
		"interval":  m.cfg.Monitor.Interval,
		"lock_file": m.lockFile,
	}
}
