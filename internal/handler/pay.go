package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/alimpay/alimpay-go/internal/config"
	"github.com/alimpay/alimpay-go/internal/database"
	"github.com/alimpay/alimpay-go/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PayHandler 支付页面处理器
type PayHandler struct {
	db  *database.DB
	cfg *config.Config
}

// NewPayHandler 创建支付页面处理器
func NewPayHandler(db *database.DB, cfg *config.Config) *PayHandler {
	return &PayHandler{
		db:  db,
		cfg: cfg,
	}
}

// HandlePayPage 处理支付页面
func (h *PayHandler) HandlePayPage(c *gin.Context) {
	tradeNo := c.Query("trade_no")
	amountStr := c.Query("amount")

	if tradeNo == "" || amountStr == "" {
		c.HTML(http.StatusOK, "error_v2.html", gin.H{
			"title":   "参数错误",
			"message": "缺少必要参数",
		})
		return
	}

	// 解析金额
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.HTML(http.StatusOK, "error_v2.html", gin.H{
			"title":   "参数错误",
			"message": "金额格式错误",
		})
		return
	}

	// 查询订单
	order, err := h.db.GetOrderByID(tradeNo)
	if err != nil || order == nil {
		c.HTML(http.StatusOK, "error_v2.html", gin.H{
			"title":   "订单不存在",
			"message": "订单未找到或已失效",
		})
		return
	}

	// 检查订单状态
	if order.Status == 1 {
		c.HTML(http.StatusOK, "error_v2.html", gin.H{
			"title":   "订单已支付",
			"message": "该订单已完成支付",
		})
		return
	}

	logger.Info("Payment page accessed",
		zap.String("trade_no", tradeNo),
		zap.Float64("amount", amount))

	// 读取经营码图片
	qrCodePath := h.cfg.Payment.BusinessQRMode.QRCodePath
	qrCodeData, err := os.ReadFile(qrCodePath)
	if err != nil {
		logger.Error("Failed to read QR code", zap.Error(err))
		c.HTML(http.StatusOK, "error_v2.html", gin.H{
			"title":   "系统错误",
			"message": "无法加载收款码",
		})
		return
	}

	// 检测文件类型
	contentType := "image/png"
	if len(qrCodeData) > 2 {
		if qrCodeData[0] == 0xFF && qrCodeData[1] == 0xD8 {
			contentType = "image/jpeg"
		}
	}

	// 渲染支付页面
	c.HTML(http.StatusOK, "pay.html", gin.H{
		"order": gin.H{
			"trade_no":       tradeNo,
			"out_trade_no":   order.OutTradeNo,
			"name":           order.Name,
			"amount":         amount,
			"payment_amount": order.PaymentAmount,
			"create_time":    order.AddTime.Format("2006-01-02 15:04:05"),
			"pid":            order.PID,
		},
		"qr_code_data": fmt.Sprintf("data:%s;base64,%s", contentType,
			encodeBase64(qrCodeData)),
		"instructions": gin.H{
			"step1": "打开支付宝，点击「扫一扫」",
			"step2": fmt.Sprintf("扫描下方二维码，输入金额 %.2f 元", amount),
			"step3": "确认支付后，页面将自动跳转",
		},
	})
}

// encodeBase64 编码为base64
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
