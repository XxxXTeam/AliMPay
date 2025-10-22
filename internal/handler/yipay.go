package handler

import (
	"net/http"
	"time"

	"alimpay-go/internal/config"
	"alimpay-go/internal/database"
	"alimpay-go/internal/model"
	"alimpay-go/internal/service"
	"alimpay-go/pkg/logger"
	"alimpay-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// YiPayHandler 易支付/码支付标准接口处理器
type YiPayHandler struct {
	db      *database.DB
	codepay *service.CodePayService
	cfg     *config.Config
}

// NewYiPayHandler 创建易支付处理器
func NewYiPayHandler(db *database.DB, codepay *service.CodePayService, cfg *config.Config) *YiPayHandler {
	return &YiPayHandler{
		db:      db,
		codepay: codepay,
		cfg:     cfg,
	}
}

// HandleMAPI 处理MAPI接口（码支付标准）
func (h *YiPayHandler) HandleMAPI(c *gin.Context) {
	// 获取act参数
	act := h.getParam(c, "act")
	if act == "" {
		act = h.getParam(c, "action")
	}

	logger.Info("MAPI request",
		zap.String("act", act),
		zap.String("ip", c.ClientIP()))

	switch act {
	case "order":
		h.HandleQueryOrder(c)
	case "orders":
		h.handleQueryOrders(c)
	default:
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Invalid act parameter",
		})
	}
}

// HandleQueryOrder 查询单个订单
func (h *YiPayHandler) HandleQueryOrder(c *gin.Context) {
	pid := h.getParam(c, "pid")
	outTradeNo := h.getParam(c, "out_trade_no")

	if pid == "" || outTradeNo == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Missing required parameters: pid, out_trade_no",
		})
		return
	}

	// 查询订单（注意参数顺序：outTradeNo, pid）
	logger.Debug("Querying order",
		zap.String("out_trade_no", outTradeNo),
		zap.String("pid", pid))

	order, err := h.db.GetOrderByOutTradeNo(outTradeNo, pid)
	if err != nil {
		logger.Error("Failed to query order",
			zap.String("out_trade_no", outTradeNo),
			zap.String("pid", pid),
			zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Order not found",
		})
		return
	}

	if order == nil {
		logger.Warn("Order is nil",
			zap.String("out_trade_no", outTradeNo),
			zap.String("pid", pid))
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Order not found",
		})
		return
	}

	logger.Info("Order found",
		zap.String("order_id", order.ID),
		zap.Int("status", order.Status))

	// 返回标准格式
	response := gin.H{
		"code":         1,
		"msg":          "SUCCESS",
		"trade_no":     order.ID,
		"out_trade_no": order.OutTradeNo,
		"type":         order.Type,
		"pid":          order.PID,
		"name":         order.Name,
		"money":        utils.FormatAmount(order.Price),
		"addtime":      order.AddTime.Format("2006-01-02 15:04:05"),
		"endtime":      "",
		"status":       order.Status, // 0=待支付, 1=已支付
	}

	if order.PayTime != nil {
		response["endtime"] = order.PayTime.Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK, response)
}

// handleQueryOrders 查询订单列表
func (h *YiPayHandler) handleQueryOrders(c *gin.Context) {
	pid := h.getParam(c, "pid")
	key := h.getParam(c, "key")

	if pid == "" || key == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Missing required parameters: pid, key",
		})
		return
	}

	// 验证商户
	merchantInfo := h.codepay.GetMerchantInfo()
	if pid != merchantInfo["id"].(string) || key != merchantInfo["key"].(string) {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Invalid merchant credentials",
		})
		return
	}

	// 获取最近订单（默认20条）
	orders, err := h.db.GetRecentOrders(20)
	if err != nil {
		logger.Error("Failed to query orders", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Failed to query orders",
		})
		return
	}

	// 构建订单列表
	orderList := make([]gin.H, 0)
	for _, order := range orders {
		item := gin.H{
			"trade_no":     order.ID,
			"out_trade_no": order.OutTradeNo,
			"type":         order.Type,
			"name":         order.Name,
			"money":        utils.FormatAmount(order.Price),
			"addtime":      order.AddTime.Format("2006-01-02 15:04:05"),
			"status":       order.Status,
		}
		if order.PayTime != nil {
			item["endtime"] = order.PayTime.Format("2006-01-02 15:04:05")
		}
		orderList = append(orderList, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   1,
		"msg":    "SUCCESS",
		"count":  len(orderList),
		"orders": orderList,
	})
}

// HandleSubmitAPI 处理API提交接口（易支付标准）
func (h *YiPayHandler) HandleSubmitAPI(c *gin.Context) {
	// 获取所有参数
	params := make(map[string]string)
	fields := []string{"pid", "type", "out_trade_no", "notify_url", "return_url",
		"name", "money", "price", "sitename", "sign", "sign_type", "param"}

	for _, field := range fields {
		params[field] = h.getParam(c, field)
	}

	// 兼容price和money
	if params["money"] == "" && params["price"] != "" {
		params["money"] = params["price"]
	}
	if params["price"] == "" && params["money"] != "" {
		params["price"] = params["money"]
	}

	if params["sign_type"] == "" {
		params["sign_type"] = "MD5"
	}

	logger.Info("Submit payment request",
		zap.String("pid", params["pid"]),
		zap.String("out_trade_no", params["out_trade_no"]),
		zap.String("money", params["money"]),
		zap.String("ip", c.ClientIP()))

	// 验证签名
	if !h.codepay.ValidateSignature(params) {
		logger.Warn("Invalid signature",
			zap.String("pid", params["pid"]),
			zap.String("out_trade_no", params["out_trade_no"]))
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "签名验证失败",
		})
		return
	}

	// 获取基础URL
	baseURL := utils.GetBaseURL(c, h.cfg.Server.BaseURL)

	// 创建订单
	result, err := h.codepay.CreatePayment(params, baseURL)
	if err != nil {
		logger.Error("Failed to create payment", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  err.Error(),
		})
		return
	}

	// 返回标准格式
	c.JSON(http.StatusOK, result)
}

// HandleClose 关闭订单
func (h *YiPayHandler) HandleClose(c *gin.Context) {
	pid := h.getParam(c, "pid")
	key := h.getParam(c, "key")
	outTradeNo := h.getParam(c, "out_trade_no")

	if pid == "" || key == "" || outTradeNo == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Missing required parameters",
		})
		return
	}

	// 验证商户
	merchantInfo := h.codepay.GetMerchantInfo()
	if pid != merchantInfo["id"].(string) || key != merchantInfo["key"].(string) {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Invalid merchant credentials",
		})
		return
	}

	// 查询订单（注意参数顺序：outTradeNo, pid）
	order, err := h.db.GetOrderByOutTradeNo(outTradeNo, pid)
	if err != nil || order == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Order not found",
		})
		return
	}

	// 检查订单状态
	if order.Status == model.OrderStatusPaid {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Order already paid, cannot close",
		})
		return
	}

	// 关闭订单
	err = h.db.UpdateOrderStatus(order.ID, model.OrderStatusClosed, time.Now())
	if err != nil {
		logger.Error("Failed to close order", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Failed to close order",
		})
		return
	}

	logger.Info("Order closed",
		zap.String("trade_no", order.ID),
		zap.String("out_trade_no", outTradeNo))

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "Order closed successfully",
	})
}

// HandleRefund 退款接口（仅返回提示）
func (h *YiPayHandler) HandleRefund(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": -1,
		"msg":  "Refund is not supported, please process manually via Alipay",
	})
}

// getParam 获取参数（支持GET和POST）
func (h *YiPayHandler) getParam(c *gin.Context, key string) string {
	value := c.Query(key)
	if value == "" {
		value = c.PostForm(key)
	}
	return value
}

// HandleQueryMerchant 查询商户信息
func (h *YiPayHandler) HandleQueryMerchant(c *gin.Context) {
	pid := h.getParam(c, "pid")
	key := h.getParam(c, "key")

	if pid == "" || key == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Missing required parameters: pid, key",
		})
		return
	}

	merchantInfo := h.codepay.GetMerchantInfo()

	if pid != merchantInfo["id"].(string) || key != merchantInfo["key"].(string) {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Invalid merchant credentials",
		})
		return
	}

	// 返回易支付标准格式
	c.JSON(http.StatusOK, gin.H{
		"code":     1,
		"pid":      merchantInfo["id"],
		"key":      utils.MaskKey(merchantInfo["key"].(string)), // 脱敏
		"active":   1,
		"money":    "0.00",
		"account":  "",
		"username": "Merchant",
		"rate":     merchantInfo["rate"],
		"issmrz":   1,
		"email":    "",
		"phone":    "",
		"url":      "",
		"addtime":  time.Now().Format("2006-01-02 15:04:05"),
	})
}

// HandleCallback 处理支付回调确认
func (h *YiPayHandler) HandleCallback(c *gin.Context) {
	// 获取参数
	params := make(map[string]string)
	fields := []string{"trade_no", "out_trade_no", "type", "name", "money",
		"trade_status", "sign", "sign_type"}

	for _, field := range fields {
		params[field] = h.getParam(c, field)
	}

	logger.Info("Received callback",
		zap.String("trade_no", params["trade_no"]),
		zap.String("out_trade_no", params["out_trade_no"]),
		zap.String("trade_status", params["trade_status"]))

	// 验证参数
	if params["trade_no"] == "" || params["out_trade_no"] == "" {
		c.String(http.StatusOK, "fail")
		return
	}

	// 验证交易状态
	if params["trade_status"] != "TRADE_SUCCESS" {
		logger.Info("Non-success trade status",
			zap.String("status", params["trade_status"]))
		c.String(http.StatusOK, "success")
		return
	}

	// 查询订单
	order, err := h.db.GetOrderByID(params["trade_no"])
	if err != nil || order == nil {
		logger.Error("Order not found",
			zap.String("trade_no", params["trade_no"]))
		c.String(http.StatusOK, "fail")
		return
	}

	// 检查是否已支付
	if order.Status == model.OrderStatusPaid {
		logger.Info("Order already paid",
			zap.String("trade_no", params["trade_no"]))
		c.String(http.StatusOK, "success")
		return
	}

	// 更新订单状态
	payTime := time.Now()
	err = h.db.UpdateOrderStatus(order.ID, model.OrderStatusPaid, payTime)
	if err != nil {
		logger.Error("Failed to update order status", zap.Error(err))
		c.String(http.StatusOK, "fail")
		return
	}

	logger.Info("Order payment confirmed",
		zap.String("trade_no", order.ID),
		zap.String("out_trade_no", order.OutTradeNo))

	// 发送商户回调
	if order.NotifyURL != "" {
		go h.codepay.SendNotification(order)
	}

	c.String(http.StatusOK, "success")
}

// HandleCheckSign 检查签名接口
func (h *YiPayHandler) HandleCheckSign(c *gin.Context) {
	// 获取所有参数
	params := make(map[string]string)

	// 从查询参数获取
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// 从POST表单获取
	c.Request.ParseForm()
	for key, values := range c.Request.PostForm {
		if len(values) > 0 && params[key] == "" {
			params[key] = values[0]
		}
	}

	// 验证签名
	valid := h.codepay.ValidateSignature(params)

	c.JSON(http.StatusOK, gin.H{
		"code": func() int {
			if valid {
				return 1
			}
			return -1
		}(),
		"msg": func() string {
			if valid {
				return "Signature valid"
			}
			return "Signature invalid"
		}(),
		"valid": valid,
	})
}
