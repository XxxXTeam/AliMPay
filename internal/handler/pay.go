package handler

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"alimpay-go/internal/config"
	"alimpay-go/internal/database"
	"alimpay-go/internal/pkg/logger"

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

	logger.Info("HandlePayPage called",
		zap.String("trade_no", tradeNo),
		zap.String("amount_str", amountStr))

	if tradeNo == "" || amountStr == "" {
		logger.Warn("Missing parameters",
			zap.String("trade_no", tradeNo),
			zap.String("amount", amountStr))
		c.HTML(http.StatusOK, "error.html", gin.H{
			"title":   "参数错误",
			"message": "缺少必要参数",
		})
		return
	}

	// 解析金额
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"title":   "参数错误",
			"message": "金额格式错误",
		})
		return
	}

	// 查询订单
	logger.Info("Querying order", zap.String("trade_no", tradeNo))
	order, err := h.db.GetOrderByID(tradeNo)
	if err != nil {
		logger.Error("Failed to query order",
			zap.String("trade_no", tradeNo),
			zap.Error(err))
		c.HTML(http.StatusOK, "error.html", gin.H{
			"title":   "订单不存在",
			"message": "订单未找到或已失效",
		})
		return
	}

	if order == nil {
		logger.Warn("Order is nil", zap.String("trade_no", tradeNo))
		c.HTML(http.StatusOK, "error.html", gin.H{
			"title":   "订单不存在",
			"message": "订单未找到或已失效",
		})
		return
	}

	logger.Info("Order found",
		zap.String("trade_no", tradeNo),
		zap.String("out_trade_no", order.OutTradeNo),
		zap.Int("status", order.Status),
		zap.Float64("payment_amount", order.PaymentAmount))

	// 检查订单状态
	if order.Status == 1 {
		logger.Warn("Order already paid", zap.String("trade_no", tradeNo))
		c.HTML(http.StatusOK, "error.html", gin.H{
			"title":   "订单已支付",
			"message": "该订单已完成支付",
		})
		return
	}

	logger.Info("Payment page accessed",
		zap.String("trade_no", tradeNo),
		zap.Float64("amount", amount))

	// 读取经营码图片
	var qrCodePath string
	var qrCodeID string

	// 如果订单有分配的二维码ID，使用对应的二维码
	if order.QRCodeID != "" && len(h.cfg.Payment.BusinessQRMode.QRCodePaths) > 0 {
		found := false
		for _, qr := range h.cfg.Payment.BusinessQRMode.QRCodePaths {
			if qr.ID == order.QRCodeID {
				qrCodePath = qr.Path
				qrCodeID = qr.CodeID
				found = true
				logger.Info("Using assigned QR code",
					zap.String("qr_id", order.QRCodeID),
					zap.String("path", qrCodePath))
				break
			}
		}
		if !found {
			logger.Warn("Assigned QR code not found, using default",
				zap.String("qr_id", order.QRCodeID))
			qrCodePath = h.cfg.Payment.BusinessQRMode.QRCodePath
			qrCodeID = h.cfg.Payment.BusinessQRMode.QRCodeID
		}
	} else {
		// 使用默认二维码
		qrCodePath = h.cfg.Payment.BusinessQRMode.QRCodePath
		qrCodeID = h.cfg.Payment.BusinessQRMode.QRCodeID
	}

	logger.Info("Reading QR code file", zap.String("path", qrCodePath))

	qrCodeData, err := os.ReadFile(qrCodePath)
	if err != nil {
		logger.Error("Failed to read QR code",
			zap.String("path", qrCodePath),
			zap.Error(err))
		c.HTML(http.StatusOK, "error.html", gin.H{
			"title":   "系统错误",
			"message": "无法加载收款码",
		})
		return
	}

	logger.Info("QR code file read successfully",
		zap.String("path", qrCodePath),
		zap.Int("size", len(qrCodeData)))

	// 检测文件类型
	contentType := "image/png"
	if len(qrCodeData) > 2 {
		if qrCodeData[0] == 0xFF && qrCodeData[1] == 0xD8 {
			contentType = "image/jpeg"
		}
	}

	// 生成 Data URI（需要使用 template.URL 类型避免被转义）
	dataURI := template.URL(fmt.Sprintf("data:%s;base64,%s", contentType,
		encodeBase64(qrCodeData)))

	logger.Info("Rendering payment page",
		zap.String("trade_no", tradeNo),
		zap.Int("qr_code_size", len(qrCodeData)))

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
		"qr_code_data": dataURI,
		"qr_code_id":   qrCodeID, // 支付宝收款码ID
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
