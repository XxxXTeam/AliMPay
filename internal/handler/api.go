package handler

import (
	"net/http"
	"strconv"

	"alimpay-go/internal/config"
	"alimpay-go/internal/service"
	"alimpay-go/internal/validator"
	"alimpay-go/pkg/logger"
	"alimpay-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// APIHandler API处理器
type APIHandler struct {
	codepay *service.CodePayService
	monitor *service.MonitorService
	cfg     *config.Config
}

// NewAPIHandler 创建API处理器
func NewAPIHandler(codepay *service.CodePayService, monitor *service.MonitorService, cfg *config.Config) *APIHandler {
	return &APIHandler{
		codepay: codepay,
		monitor: monitor,
		cfg:     cfg,
	}
}

// HandleAction 处理API请求
func (h *APIHandler) HandleAction(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		action = c.PostForm("action")
	}
	if action == "" {
		action = c.Query("act") // 支持易支付的act参数
	}
	if action == "" {
		action = c.PostForm("act")
	}

	if action == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": -1,
			"msg":  "Missing action parameter",
		})
		return
	}

	logger.Info("API request", zap.String("action", action), zap.String("ip", c.ClientIP()))

	switch action {
	case "query":
		h.handleQueryMerchant(c)
	case "order":
		h.handleQueryOrder(c)
	case "orders":
		h.handleQueryOrders(c)
	case "submit", "create":
		h.handleCreatePayment(c)
	case "health":
		h.handleHealth(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code": -1,
			"msg":  "Invalid action",
		})
	}
}

// handleQueryMerchant 查询商户信息
func (h *APIHandler) handleQueryMerchant(c *gin.Context) {
	pid := h.getParam(c, "pid")
	key := h.getParam(c, "key")

	if pid == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": -1,
			"msg":  "Missing required parameters: pid, key",
		})
		return
	}

	merchantInfo := h.codepay.GetMerchantInfo()

	if pid != merchantInfo["id"] || key != merchantInfo["key"] {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Invalid merchant credentials",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     1,
		"pid":      merchantInfo["id"],
		"key":      utils.MaskKey(merchantInfo["key"].(string)), // 脱敏处理
		"qq":       nil,
		"active":   1,
		"money":    "0.00",
		"account":  "",
		"username": "Merchant",
		"rate":     merchantInfo["rate"],
		"issmrz":   1,
	})
}

// handleQueryOrder 查询单个订单
func (h *APIHandler) handleQueryOrder(c *gin.Context) {
	pid := h.getParam(c, "pid")
	key := h.getParam(c, "key")
	outTradeNo := h.getParam(c, "out_trade_no")

	if pid == "" || outTradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": -1,
			"msg":  "Missing required parameters",
		})
		return
	}

	// 允许不验证key的查询（用于前端状态检查）
	validateKey := key != ""
	result, err := h.codepay.QueryOrder(pid, key, outTradeNo, validateKey)
	if err != nil {
		logger.Error("Failed to query order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleQueryOrders 查询订单列表
func (h *APIHandler) handleQueryOrders(c *gin.Context) {
	pid := h.getParam(c, "pid")
	key := h.getParam(c, "key")
	limitStr := h.getParam(c, "limit")

	if pid == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": -1,
			"msg":  "Missing required parameters: pid, key",
		})
		return
	}

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	result, err := h.codepay.QueryOrders(pid, key, limit)
	if err != nil {
		logger.Error("Failed to query orders", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleCreatePayment 创建支付
func (h *APIHandler) handleCreatePayment(c *gin.Context) {
	params := make(map[string]string)

	// 获取所有参数（兼容易支付：不限制参数字段）
	// 从 Query 参数获取
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// 从 POST 表单获取（如果存在则覆盖）
	if c.Request.Method == "POST" {
		if err := c.Request.ParseForm(); err != nil {
			logger.Error("Failed to parse form", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "Invalid form data"})
			return
		}
		for key, values := range c.Request.PostForm {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
	}

	// 兼容易支付：如果没有money但有price，复制price到money
	if params["money"] == "" && params["price"] != "" {
		params["money"] = params["price"]
	}

	if params["sign_type"] == "" {
		params["sign_type"] = "MD5"
	}

	// 验证参数
	if err := validator.ValidateOrderParams(params); err != nil {
		logger.Warn("Invalid order parameters", zap.Error(err), zap.String("out_trade_no", params["out_trade_no"]))
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	// 验证签名（防止伪造请求和0元购）
	if !h.codepay.ValidateSignature(params) {
		logger.Warn("Invalid signature",
			zap.String("pid", params["pid"]),
			zap.String("out_trade_no", params["out_trade_no"]),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": -1,
			"msg":  "签名验证失败",
		})
		return
	}

	// 获取基础URL
	baseURL := utils.GetBaseURL(c, h.cfg.Server.BaseURL)

	result, err := h.codepay.CreatePayment(params, baseURL)
	if err != nil {
		logger.Error("Failed to create payment", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code": -1,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleHealth 健康检查
func (h *APIHandler) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":      1,
		"msg":       "System is healthy",
		"timestamp": "2006-01-02 15:04:05",
	})
}

// getParam 获取参数（支持GET和POST）
func (h *APIHandler) getParam(c *gin.Context, key string) string {
	value := c.Query(key)
	if value == "" {
		value = c.PostForm(key)
	}
	return value
}
