package handler

import (
	"net/http"

	"github.com/alimpay/alimpay-go/internal/service"
	"github.com/alimpay/alimpay-go/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NotifyHandler 支付通知处理器
type NotifyHandler struct {
	codepay *service.CodePayService
}

// NewNotifyHandler 创建通知处理器
func NewNotifyHandler(codepay *service.CodePayService) *NotifyHandler {
	return &NotifyHandler{
		codepay: codepay,
	}
}

// HandleNotify 处理支付宝异步通知
func (h *NotifyHandler) HandleNotify(c *gin.Context) {
	// 获取所有参数
	params := make(map[string]string)

	// 支持GET和POST
	if c.Request.Method == "POST" {
		c.Request.ParseForm()
		for key, values := range c.Request.PostForm {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}
	}

	// 从URL查询参数获取
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 && params[key] == "" {
			params[key] = values[0]
		}
	}

	logger.Info("Received payment notification",
		zap.String("method", c.Request.Method),
		zap.Int("param_count", len(params)),
		zap.String("remote_addr", c.ClientIP()))

	// 验证必需参数
	tradeNo := params["trade_no"]
	outTradeNo := params["out_trade_no"]
	tradeStatus := params["trade_status"]

	if tradeNo == "" || outTradeNo == "" {
		logger.Warn("Missing required parameters in notification")
		c.String(http.StatusOK, "fail")
		return
	}

	// 检查交易状态
	if tradeStatus != "TRADE_SUCCESS" {
		logger.Info("Non-success trade status",
			zap.String("trade_no", tradeNo),
			zap.String("status", tradeStatus))
		c.String(http.StatusOK, "success") // 仍返回success表示已接收
		return
	}

	logger.Info("Payment notification processed successfully",
		zap.String("trade_no", tradeNo),
		zap.String("out_trade_no", outTradeNo),
		zap.String("trade_status", tradeStatus))

	c.String(http.StatusOK, "success")
}
