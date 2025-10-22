package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alimpay/alimpay-go/internal/config"
	"github.com/alimpay/alimpay-go/internal/database"
	"github.com/alimpay/alimpay-go/internal/model"
	"github.com/alimpay/alimpay-go/pkg/lock"
	"github.com/alimpay/alimpay-go/pkg/logger"
	"github.com/alimpay/alimpay-go/pkg/qrcode"
	"github.com/alimpay/alimpay-go/pkg/utils"
	"go.uber.org/zap"
)

// CodePayService 码支付服务
type CodePayService struct {
	cfg          *config.Config
	db           *database.DB
	transfer     *AlipayTransfer
	qrGenerator  *qrcode.Generator
	merchantID   string
	alipayClient *AlipayClient
	merchantKey  string
}

// NewCodePayService 创建码支付服务
func NewCodePayService(cfg *config.Config, db *database.DB) (*CodePayService, error) {
	// 创建支付宝客户端
	alipayClient, err := NewAlipayClient(&cfg.Alipay)
	if err != nil {
		return nil, fmt.Errorf("failed to create alipay client: %w", err)
	}

	service := &CodePayService{
		cfg:          cfg,
		db:           db,
		transfer:     NewAlipayTransfer(&cfg.Alipay),
		qrGenerator:  qrcode.NewGenerator(cfg.Payment.QRCodeSize, cfg.Payment.QRCodeMargin),
		alipayClient: alipayClient,
	}

	// 初始化商户信息
	if err := service.initMerchant(); err != nil {
		return nil, err
	}

	return service, nil
}

// initMerchant 初始化商户信息
func (s *CodePayService) initMerchant() error {
	if s.cfg.Merchant.ID != "" && s.cfg.Merchant.Key != "" {
		s.merchantID = s.cfg.Merchant.ID
		s.merchantKey = s.cfg.Merchant.Key
		logger.Info("Loaded merchant configuration",
			zap.String("merchant_id", s.merchantID))
		return nil
	}

	// 生成新的商户信息
	s.merchantID = utils.GenerateMerchantID()
	s.merchantKey = utils.GenerateMerchantKey()

	// 保存到配置
	s.cfg.Merchant.ID = s.merchantID
	s.cfg.Merchant.Key = s.merchantKey

	// 保存配置文件
	configPath := "./configs/config.yaml"
	if err := config.Save(s.cfg, configPath); err != nil {
		logger.Warn("Failed to save merchant config", zap.Error(err))
	}

	logger.Info("Generated new merchant configuration",
		zap.String("merchant_id", s.merchantID),
		zap.String("merchant_key", s.merchantKey))

	return nil
}

// GetMerchantInfo 获取商户信息
func (s *CodePayService) GetMerchantInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":   s.merchantID,
		"key":  s.merchantKey,
		"rate": s.cfg.Merchant.Rate,
	}
}

// CreatePayment 创建支付订单
func (s *CodePayService) CreatePayment(params map[string]string) (map[string]interface{}, error) {
	// 验证参数
	if err := s.validatePaymentParams(params); err != nil {
		return nil, err
	}

	// 验证签名
	if !utils.VerifySign(params, s.merchantKey) {
		return nil, fmt.Errorf("invalid signature")
	}

	// 检查订单是否已存在（防止重复提交）
	existingOrder, err := s.db.GetOrderByOutTradeNo(params["out_trade_no"], params["pid"])
	if err != nil {
		return nil, fmt.Errorf("failed to check existing order: %w", err)
	}

	// 如果订单已存在，返回已有订单信息
	if existingOrder != nil {
		logger.Info("Order already exists, returning existing order",
			zap.String("out_trade_no", params["out_trade_no"]),
			zap.String("trade_no", existingOrder.ID))
		return s.buildOrderResponse(existingOrder), nil
	}

	// 解析金额（严格防止0元购）
	var amount float64
	moneyStr := params["money"]
	if moneyStr == "" {
		moneyStr = params["price"] // 兼容price参数
	}

	_, err = fmt.Sscanf(moneyStr, "%f", &amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	// 严格验证金额（防止0元购）
	if amount <= 0 {
		return nil, fmt.Errorf("invalid amount: must be greater than 0 (0 yuan purchase not allowed)")
	}

	if amount < 0.01 {
		return nil, fmt.Errorf("invalid amount: minimum is 0.01 yuan")
	}

	if amount > 99999.99 {
		return nil, fmt.Errorf("invalid amount: maximum is 99999.99 yuan")
	}

	// 生成交易号
	tradeNo := utils.GenerateTradeNo()

	// 确定支付金额（经营码模式可能需要调整）
	paymentAmount := amount
	amountAdjusted := false
	adjustmentNote := ""

	if s.cfg.Payment.BusinessQRMode.Enabled {
		var err error
		paymentAmount, err = s.allocateUniqueAmount(amount)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate unique amount: %w", err)
		}

		if paymentAmount != amount {
			amountAdjusted = true
			adjustmentNote = fmt.Sprintf("检测到相同金额订单，实际支付金额已调整为 %.2f 元", paymentAmount)
		}
	}

	// 创建订单
	order := &model.Order{
		ID:            tradeNo,
		OutTradeNo:    params["out_trade_no"],
		Type:          params["type"],
		PID:           params["pid"],
		Name:          params["name"],
		Price:         amount,
		PaymentAmount: paymentAmount,
		Status:        model.OrderStatusPending,
		AddTime:       time.Now(),
		NotifyURL:     params["notify_url"],
		ReturnURL:     params["return_url"],
		Sitename:      params["sitename"],
	}

	if err := s.db.CreateOrder(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	logger.Info("Order created",
		zap.String("trade_no", tradeNo),
		zap.String("out_trade_no", params["out_trade_no"]),
		zap.Float64("amount", amount),
		zap.Float64("payment_amount", paymentAmount))

	// 注意：本系统使用账单查询方式监听支付（和PHP版本一致）
	// 不需要 alipay.trade.query 接口权限
	// 监听服务会每30秒自动查询账单并匹配订单

	// 生成支付信息
	response := map[string]interface{}{
		"code":           1,
		"msg":            "SUCCESS",
		"pid":            params["pid"],
		"trade_no":       tradeNo,
		"out_trade_no":   params["out_trade_no"],
		"money":          utils.FormatAmount(amount),
		"payment_amount": paymentAmount,
		"create_time":    order.AddTime.Format("2006-01-02 15:04:05"), // 订单创建时间
	}

	// 根据收款模式生成二维码
	if s.cfg.Payment.BusinessQRMode.Enabled {
		// 经营码模式：生成包含金额信息的支付链接
		baseURL := s.getBaseURL()

		// 生成支付页面链接（包含金额信息）
		paymentPageURL := fmt.Sprintf("%s/pay?trade_no=%s&amount=%.2f",
			baseURL, tradeNo, paymentAmount)

		// 生成二维码（用户扫码后跳转到支付页面）
		qrCodeBase64, err := s.qrGenerator.GenerateToBase64(paymentPageURL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate QR code: %w", err)
		}

		response["payment_url"] = paymentPageURL
		response["qr_code"] = qrCodeBase64
		response["business_qr_mode"] = true
		response["payment_instruction"] = fmt.Sprintf("请使用支付宝扫描二维码，确认支付 %.2f 元", paymentAmount)

		if amountAdjusted {
			response["amount_adjusted"] = true
			response["adjustment_note"] = adjustmentNote
			response["original_amount"] = amount
		}

		response["payment_tips"] = []string{
			fmt.Sprintf("请务必支付准确金额：%.2f 元", paymentAmount),
			"支付时无需填写备注信息",
			"请在5分钟内完成支付，超时订单将被自动删除",
			"支付完成后系统会自动检测到账",
			"如长时间未到账，请联系客服",
		}

	} else {
		// 传统转账模式：生成动态转账二维码
		transferURL := s.transfer.GenerateTransferURL(paymentAmount, params["out_trade_no"], "")
		qrCodeBase64, err := s.qrGenerator.GenerateToBase64(transferURL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate QR code: %w", err)
		}

		response["payment_url"] = transferURL
		response["qr_code"] = qrCodeBase64
	}

	return response, nil
}

// buildOrderResponse 构建订单响应（用于已存在的订单）
func (s *CodePayService) buildOrderResponse(order *model.Order) map[string]interface{} {
	response := map[string]interface{}{
		"code":           1,
		"msg":            "SUCCESS",
		"pid":            order.PID,
		"trade_no":       order.ID,
		"out_trade_no":   order.OutTradeNo,
		"money":          utils.FormatAmount(order.Price),
		"payment_amount": order.PaymentAmount,
		"create_time":    order.AddTime.Format("2006-01-02 15:04:05"), // 订单创建时间
	}

	// 根据收款模式生成二维码
	if s.cfg.Payment.BusinessQRMode.Enabled {
		// 经营码模式
		token := utils.MD5(fmt.Sprintf("qrcode_access_%s", time.Now().Format("2006-01-02")))
		baseURL := s.getBaseURL()
		qrCodeURL := fmt.Sprintf("%s/qrcode?type=business&token=%s", baseURL, token)

		response["payment_url"] = "" // 经营码模式没有直接URL
		response["qr_code_url"] = qrCodeURL
		response["business_qr_mode"] = true
		response["payment_instruction"] = fmt.Sprintf("请使用支付宝扫描二维码，支付金额：%.2f 元", order.PaymentAmount)

		// 检查金额是否被调整
		if order.PaymentAmount != order.Price {
			response["amount_adjusted"] = true
			response["adjustment_note"] = fmt.Sprintf("检测到相同金额订单，实际支付金额已调整为 %.2f 元", order.PaymentAmount)
			response["original_amount"] = order.Price
		}

		response["payment_tips"] = []string{
			fmt.Sprintf("请务必支付准确金额：%.2f 元", order.PaymentAmount),
			"支付时无需填写备注信息",
			"请在5分钟内完成支付，超时订单将被自动删除",
			"支付完成后系统会自动检测到账",
			"如长时间未到账，请联系客服",
		}
	} else {
		// 传统转账模式
		transferURL := s.transfer.GenerateTransferURL(order.PaymentAmount, order.OutTradeNo, "")
		qrCodeBase64, _ := s.qrGenerator.GenerateToBase64(transferURL)

		response["payment_url"] = transferURL
		response["qr_code"] = qrCodeBase64
	}

	return response
}

// allocateUniqueAmount 分配唯一的支付金额
func (s *CodePayService) allocateUniqueAmount(originalAmount float64) (float64, error) {
	amountLock := lock.GetAmountLock()
	amountLock.Lock()
	defer amountLock.Unlock()

	offset := s.cfg.Payment.BusinessQRMode.AmountOffset
	timeout := s.cfg.Payment.OrderTimeout
	sinceTime := time.Now().Add(-time.Duration(timeout) * time.Second)

	paymentAmount := originalAmount
	maxAttempts := 100

	for i := 0; i < maxAttempts; i++ {
		exists, err := s.db.CheckAmountExists(paymentAmount, sinceTime)
		if err != nil {
			return 0, err
		}

		if !exists {
			logger.Info("Unique amount allocated",
				zap.Float64("original", originalAmount),
				zap.Float64("allocated", paymentAmount),
				zap.Int("attempts", i+1))
			return paymentAmount, nil
		}

		paymentAmount += offset
	}

	return 0, fmt.Errorf("failed to allocate unique amount after %d attempts", maxAttempts)
}

// QueryOrder 查询订单
func (s *CodePayService) QueryOrder(pid, key, outTradeNo string, validateKey bool) (map[string]interface{}, error) {
	if validateKey && (pid != s.merchantID || key != s.merchantKey) {
		return map[string]interface{}{
			"code": -1,
			"msg":  "Invalid merchant credentials",
		}, nil
	}

	if !validateKey && pid != s.merchantID {
		return map[string]interface{}{
			"code": -1,
			"msg":  "Invalid merchant ID",
		}, nil
	}

	order, err := s.db.GetOrderByOutTradeNo(outTradeNo, pid)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return map[string]interface{}{
			"code": -1,
			"msg":  "Order not found",
		}, nil
	}

	return map[string]interface{}{
		"code":         1,
		"msg":          "SUCCESS",
		"trade_no":     order.ID,
		"out_trade_no": order.OutTradeNo,
		"type":         order.Type,
		"pid":          order.PID,
		"addtime":      utils.FormatTime(order.AddTime),
		"endtime":      s.formatPayTime(order.PayTime),
		"name":         order.Name,
		"money":        utils.FormatAmount(order.Price),
		"status":       order.Status,
	}, nil
}

// QueryOrders 查询订单列表
func (s *CodePayService) QueryOrders(pid, key string, limit int) ([]map[string]interface{}, error) {
	if pid != s.merchantID || key != s.merchantKey {
		return nil, fmt.Errorf("invalid merchant credentials")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	orders, err := s.db.GetOrders(pid, limit)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, order := range orders {
		result = append(result, map[string]interface{}{
			"trade_no":     order.ID,
			"out_trade_no": order.OutTradeNo,
			"type":         order.Type,
			"pid":          order.PID,
			"addtime":      utils.FormatTime(order.AddTime),
			"endtime":      s.formatPayTime(order.PayTime),
			"name":         order.Name,
			"money":        utils.FormatAmount(order.Price),
			"status":       order.Status,
		})
	}

	return result, nil
}

// validatePaymentParams 验证支付参数
func (s *CodePayService) validatePaymentParams(params map[string]string) error {
	required := []string{"pid", "type", "out_trade_no", "notify_url", "return_url", "name", "money", "sign"}
	for _, field := range required {
		if params[field] == "" {
			return fmt.Errorf("missing required parameter: %s", field)
		}
	}

	if params["pid"] != s.merchantID {
		return fmt.Errorf("invalid merchant ID")
	}

	if params["type"] != model.PaymentTypeAlipay {
		return fmt.Errorf("only alipay payment type is supported")
	}

	return nil
}

// formatPayTime 格式化支付时间
func (s *CodePayService) formatPayTime(payTime *time.Time) string {
	if payTime == nil {
		return ""
	}
	return utils.FormatTime(*payTime)
}

// getBaseURL 获取基础URL
func (s *CodePayService) getBaseURL() string {
	// 在实际环境中，这应该从请求中获取
	// 这里简化处理，返回配置的服务器地址
	return fmt.Sprintf("http://localhost:%d", s.cfg.Server.Port)
}

// GetMerchantID 获取商户ID
func (s *CodePayService) GetMerchantID() string {
	return s.merchantID
}

// GetMerchantKey 获取商户密钥
func (s *CodePayService) GetMerchantKey() string {
	return s.merchantKey
}

// SendNotification 发送支付通知给商户
func (s *CodePayService) SendNotification(order *model.Order) error {
	if order.NotifyURL == "" {
		logger.Warn("No notify URL configured", zap.String("order_id", order.ID))
		return nil
	}

	notifyData := map[string]string{
		"pid":          order.PID,
		"trade_no":     order.ID,
		"out_trade_no": order.OutTradeNo,
		"type":         order.Type,
		"name":         order.Name,
		"money":        utils.FormatAmount(order.Price),
		"trade_status": "TRADE_SUCCESS",
	}

	// 生成签名
	sign := utils.GenerateSign(notifyData, s.merchantKey)
	notifyData["sign"] = sign
	notifyData["sign_type"] = "MD5"

	logger.Info("Sending notification to merchant",
		zap.String("order_id", order.ID),
		zap.String("out_trade_no", order.OutTradeNo),
		zap.String("notify_url", order.NotifyURL),
		zap.String("sign", utils.MaskSign(sign))) // 签名脱敏

	// 实际发送HTTP通知
	return s.sendHTTPNotification(order.NotifyURL, notifyData)
}

// ProcessPaymentCallback 处理支付回调（内部使用）
func (s *CodePayService) ProcessPaymentCallback(tradeNo string, paymentAmount float64, billTime string) error {
	// 查询订单
	order, err := s.db.GetOrderByID(tradeNo)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return fmt.Errorf("order not found: %s", tradeNo)
	}

	// 检查订单状态
	if order.Status == model.OrderStatusPaid {
		logger.Info("Order already paid", zap.String("trade_no", tradeNo))
		return nil
	}

	// 验证金额
	if order.PaymentAmount != paymentAmount {
		return fmt.Errorf("payment amount mismatch: expected %.2f, got %.2f",
			order.PaymentAmount, paymentAmount)
	}

	// 更新订单状态
	payTime := time.Now()
	if err := s.db.UpdateOrderStatus(order.ID, model.OrderStatusPaid, payTime); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	logger.Info("Order payment confirmed",
		zap.String("trade_no", tradeNo),
		zap.String("out_trade_no", order.OutTradeNo),
		zap.Float64("amount", paymentAmount))

	// 发送通知给商户
	if err := s.SendNotification(order); err != nil {
		logger.Error("Failed to send merchant notification",
			zap.String("trade_no", tradeNo),
			zap.Error(err))
		// 不返回错误，因为订单已经更新成功
	}

	return nil
}

// sendHTTPNotification 发送HTTP通知
func (s *CodePayService) sendHTTPNotification(notifyURL string, data map[string]string) error {
	// 构建查询字符串
	values := make(url.Values)
	for k, v := range data {
		values.Add(k, v)
	}

	// 拼接完整URL
	fullURL := notifyURL
	if strings.Contains(notifyURL, "?") {
		fullURL += "&" + values.Encode()
	} else {
		fullURL += "?" + values.Encode()
	}

	// 创建HTTP客户端（设置超时）
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送GET请求
	resp, err := client.Get(fullURL)
	if err != nil {
		logger.Error("Failed to send notification", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read notification response", zap.Error(err))
		return err
	}

	responseStr := string(body)
	responseLower := strings.TrimSpace(strings.ToLower(responseStr))

	// 检查响应是否为 "success" 或 "ok"
	if responseLower == "success" || responseLower == "ok" {
		logger.Info("Notification sent successfully",
			zap.String("notify_url", notifyURL),
			zap.String("response", responseStr))
		return nil
	}

	// 如果是测试URL（example.com），不报错，只记录警告
	if strings.Contains(notifyURL, "example.com") {
		logger.Warn("Test notify URL, skipping validation",
			zap.String("notify_url", notifyURL),
			zap.String("response_preview", responseStr[:min(len(responseStr), 100)]+"..."))
		return nil // 测试URL不报错
	}

	logger.Warn("Notification response is not success",
		zap.String("notify_url", notifyURL),
		zap.String("response", responseStr))

	return fmt.Errorf("invalid notification response: %s", responseStr)
}

// CleanupExpiredOrders 清理过期订单
func (s *CodePayService) CleanupExpiredOrders() (int64, error) {
	if !s.cfg.Payment.AutoCleanup {
		return 0, nil
	}

	timeout := s.cfg.Payment.OrderTimeout
	expiredTime := time.Now().Add(-time.Duration(timeout) * time.Second)

	count, err := s.db.DeleteExpiredOrders(expiredTime)
	if err != nil {
		return 0, err
	}

	if count > 0 {
		logger.Info("Cleaned up expired orders",
			zap.Int64("count", count),
			zap.String("expired_before", utils.FormatTime(expiredTime)))
	}

	return count, nil
}
