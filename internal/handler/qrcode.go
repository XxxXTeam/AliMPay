package handler

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"alimpay-go/internal/config"
	"alimpay-go/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// QRCodeHandler 二维码处理器
type QRCodeHandler struct {
	cfg *config.Config
}

// NewQRCodeHandler 创建二维码处理器
func NewQRCodeHandler(cfg *config.Config) *QRCodeHandler {
	return &QRCodeHandler{
		cfg: cfg,
	}
}

// HandleQRCode 处理二维码请求
func (h *QRCodeHandler) HandleQRCode(c *gin.Context) {
	qrType := c.Query("type")
	token := c.Query("token")
	qrID := c.Query("id") // 二维码ID（多二维码模式）

	// 验证token
	expectedToken := h.generateToken()
	if token != expectedToken {
		c.String(http.StatusForbidden, "Invalid token")
		return
	}

	switch qrType {
	case "business":
		h.handleBusinessQRCode(c, qrID)
	default:
		c.String(http.StatusBadRequest, "Invalid QR code type")
	}
}

// handleBusinessQRCode 处理经营码二维码
func (h *QRCodeHandler) handleBusinessQRCode(c *gin.Context, qrID string) {
	var qrCodePath string

	// 如果配置了多个二维码
	if len(h.cfg.Payment.BusinessQRMode.QRCodePaths) > 0 {
		if qrID == "" {
			// 未指定ID，使用第一个
			qrCodePath = h.cfg.Payment.BusinessQRMode.QRCodePaths[0].Path
		} else {
			// 根据ID查找对应的二维码
			found := false
			for _, qr := range h.cfg.Payment.BusinessQRMode.QRCodePaths {
				if qr.ID == qrID {
					qrCodePath = qr.Path
					found = true
					break
				}
			}
			if !found {
				logger.Error("QR code not found", zap.String("id", qrID))
				c.String(http.StatusNotFound, "QR code not found")
				return
			}
		}
	} else {
		// 传统单二维码模式
		qrCodePath = h.cfg.Payment.BusinessQRMode.QRCodePath
	}

	// 检查文件是否存在
	if _, err := os.Stat(qrCodePath); os.IsNotExist(err) {
		logger.Error("Business QR code file not found", zap.String("path", qrCodePath))
		c.String(http.StatusNotFound, "Business QR code file not found")
		return
	}

	// 读取文件
	data, err := os.ReadFile(qrCodePath)
	if err != nil {
		logger.Error("Failed to read QR code file", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to read QR code file")
		return
	}

	// 设置响应头
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=3600")

	// 返回文件
	c.Data(http.StatusOK, "image/png", data)
}

// generateToken 生成访问token
func (h *QRCodeHandler) generateToken() string {
	data := fmt.Sprintf("qrcode_access_%s", time.Now().Format("2006-01-02"))
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
