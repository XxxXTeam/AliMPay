package handler

import (
	"net/http"

	"alimpay-go/internal/config"
	"alimpay-go/internal/service"
	"alimpay-go/pkg/logger"
	"alimpay-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SubmitHandler 支付页面处理器
type SubmitHandler struct {
	codepay *service.CodePayService
	cfg     *config.Config
}

// NewSubmitHandler 创建支付页面处理器
func NewSubmitHandler(codepay *service.CodePayService, cfg *config.Config) *SubmitHandler {
	return &SubmitHandler{
		codepay: codepay,
		cfg:     cfg,
	}
}

// HandleSubmit 处理支付页面请求
func (h *SubmitHandler) HandleSubmit(c *gin.Context) {
	// 获取所有参数（兼容易支付：不限制参数字段）
	params := make(map[string]string)

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

	// 设置默认签名类型
	if params["sign_type"] == "" {
		params["sign_type"] = "MD5"
	}

	// 获取基础URL
	baseURL := utils.GetBaseURL(c, h.cfg.Server.BaseURL)

	// 创建支付
	result, err := h.codepay.CreatePayment(params, baseURL)
	if err != nil {
		logger.Error("Failed to create payment", zap.Error(err))
		h.renderError(c, err.Error())
		return
	}

	// 渲染支付页面
	h.renderPaymentPage(c, result, params)
}

// renderPaymentPage 渲染支付页面
func (h *SubmitHandler) renderPaymentPage(c *gin.Context, result map[string]interface{}, params map[string]string) {
	// 准备模板数据
	templateData := gin.H{
		// 基本信息
		"PID":        params["pid"],
		"OutTradeNo": params["out_trade_no"],
		"Name":       params["name"],
		"SiteName":   params["sitename"],
		"ReturnURL":  params["return_url"],

		// 支付信息
		"TradeNo":       getString(result, "trade_no"),
		"PaymentAmount": getFloat(result, "payment_amount"),
		"PaymentURL":    getString(result, "payment_url"),
		"QrCode":        getString(result, "qr_code"),
		"QrCodeURL":     getString(result, "qr_code_url"),
		"QRCodeID":      h.cfg.Payment.BusinessQRMode.QRCodeID, // 支付宝收款码ID（用于拉起APP）
		"CreateTime":    getString(result, "create_time"),      // 订单创建时间

		// 模式和提示
		"BusinessQrMode": getBool(result, "business_qr_mode"),
		"AmountAdjusted": getBool(result, "amount_adjusted"),
		"AdjustmentNote": getString(result, "adjustment_note"),
		"PaymentTips":    getSlice(result, "payment_tips"),
	}

	// 渲染模板
	c.HTML(http.StatusOK, "submit.html", templateData)
}

// 辅助函数：安全获取字符串
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// 辅助函数：安全获取布尔值
func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// 辅助函数：安全获取浮点数
func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0.0
}

// 辅助函数：安全获取切片
func getSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		if slice, ok := v.([]string); ok {
			return slice
		}
	}
	return []string{}
}

// renderError 渲染错误页面
func (h *SubmitHandler) renderError(c *gin.Context, errorMsg string) {
	c.HTML(http.StatusOK, "error.html", gin.H{
		"error": errorMsg,
	})
}
