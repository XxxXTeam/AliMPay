package handler

import (
	"net/http"
	"strconv"

	"alimpay-go/internal/config"
	"alimpay-go/pkg/logger"
	"alimpay-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AlipayLinkHandler 支付宝深链接处理器
type AlipayLinkHandler struct {
	cfg *config.Config
}

// NewAlipayLinkHandler 创建支付宝深链接处理器
func NewAlipayLinkHandler(cfg *config.Config) *AlipayLinkHandler {
	return &AlipayLinkHandler{
		cfg: cfg,
	}
}

// HandleGenerateLink 生成支付宝深链接
// 支持两种模式：
// 1. 使用配置的qrCodeId: GET /alipay/link?amount=1.00&remark=备注
// 2. 自定义qrCodeId: GET /alipay/link?qr_code_id=xxx&amount=1.00&remark=备注
func (h *AlipayLinkHandler) HandleGenerateLink(c *gin.Context) {
	// 获取参数
	qrCodeID := c.Query("qr_code_id")
	amountStr := c.Query("amount")
	remark := c.Query("remark")

	// 如果没有提供qrCodeId，使用配置中的
	if qrCodeID == "" {
		qrCodeID = h.cfg.Payment.BusinessQRMode.QRCodeID
	}

	// 验证qrCodeId
	if qrCodeID == "" {
		logger.Warn("Missing qr_code_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  -1,
			"msg":   "缺少二维码ID参数",
			"error": "qr_code_id is required",
		})
		return
	}

	// 解析金额
	var amount float64
	if amountStr != "" {
		var err error
		amount, err = strconv.ParseFloat(amountStr, 64)
		if err != nil {
			logger.Warn("Invalid amount parameter", zap.String("amount", amountStr))
			c.JSON(http.StatusBadRequest, gin.H{
				"code":  -1,
				"msg":   "金额格式错误",
				"error": "invalid amount format",
			})
			return
		}

		// 验证金额范围
		if amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":  -1,
				"msg":   "金额必须大于0",
				"error": "amount must be greater than 0",
			})
			return
		}

		if amount > 99999.99 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":  -1,
				"msg":   "金额不能超过99999.99元",
				"error": "amount exceeds maximum limit",
			})
			return
		}
	}

	// 生成深链接
	deepLink := utils.GenerateAlipayDeepLink(qrCodeID, amount, remark)

	logger.Info("Generated Alipay deep link",
		zap.String("qr_code_id", qrCodeID),
		zap.Float64("amount", amount),
		zap.String("remark", remark))

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":            1,
		"msg":             "SUCCESS",
		"qr_code_id":      qrCodeID,
		"amount":          amount,
		"remark":          remark,
		"alipay_deep_link": deepLink,
		"usage":           "在移动端浏览器中访问此链接可直接拉起支付宝进行支付",
	})
}

// HandleRedirectToAlipay 直接重定向到支付宝
// GET /alipay/pay?amount=1.00&remark=备注
func (h *AlipayLinkHandler) HandleRedirectToAlipay(c *gin.Context) {
	// 获取参数
	qrCodeID := c.Query("qr_code_id")
	amountStr := c.Query("amount")
	remark := c.Query("remark")

	// 如果没有提供qrCodeId，使用配置中的
	if qrCodeID == "" {
		qrCodeID = h.cfg.Payment.BusinessQRMode.QRCodeID
	}

	// 验证qrCodeId
	if qrCodeID == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"title":   "参数错误",
			"message": "未配置支付宝二维码ID",
		})
		return
	}

	// 解析金额
	var amount float64
	if amountStr != "" {
		var err error
		amount, err = strconv.ParseFloat(amountStr, 64)
		if err != nil || amount <= 0 {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"title":   "参数错误",
				"message": "金额格式错误或金额必须大于0",
			})
			return
		}
	}

	// 生成深链接
	deepLink := utils.GenerateAlipayDeepLink(qrCodeID, amount, remark)

	logger.Info("Redirecting to Alipay",
		zap.String("qr_code_id", qrCodeID),
		zap.Float64("amount", amount),
		zap.String("remark", remark))

	// 重定向到支付宝
	c.Redirect(http.StatusFound, deepLink)
}
