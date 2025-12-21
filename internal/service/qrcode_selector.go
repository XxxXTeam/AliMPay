package service

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"alimpay-go/internal/config"
	"alimpay-go/internal/pkg/logger"

	"go.uber.org/zap"
)

// QRCodeSelector 二维码选择器
// @description 负责选择和分配二维码给订单
type QRCodeSelector struct {
	cfg          *config.Config
	qrCodes      []config.QRCode
	currentIndex int
	usageCount   map[string]int
	lastUsedTime map[string]time.Time
	mu           sync.RWMutex
	pollingMode  string
}

// NewQRCodeSelector 创建二维码选择器
func NewQRCodeSelector(cfg *config.Config) *QRCodeSelector {
	// 过滤出启用的二维码并按优先级排序
	var enabledQRCodes []config.QRCode
	for _, qr := range cfg.Payment.BusinessQRMode.QRCodePaths {
		if qr.Enabled {
			enabledQRCodes = append(enabledQRCodes, qr)
		}
	}

	// 如果没有配置多个二维码，返回nil（使用传统单二维码模式）
	if len(enabledQRCodes) == 0 {
		logger.Warn("No enabled QR codes found, falling back to single QR code mode")
		return nil
	}

	// 按优先级排序（优先级数字越小越高）
	for i := 0; i < len(enabledQRCodes)-1; i++ {
		for j := i + 1; j < len(enabledQRCodes); j++ {
			if enabledQRCodes[i].Priority > enabledQRCodes[j].Priority {
				enabledQRCodes[i], enabledQRCodes[j] = enabledQRCodes[j], enabledQRCodes[i]
			}
		}
	}

	pollingMode := cfg.Payment.BusinessQRMode.PollingMode
	if pollingMode == "" {
		pollingMode = "round_robin"
	}

	selector := &QRCodeSelector{
		cfg:          cfg,
		qrCodes:      enabledQRCodes,
		currentIndex: 0,
		usageCount:   make(map[string]int),
		lastUsedTime: make(map[string]time.Time),
		pollingMode:  pollingMode,
	}

	logger.Info("QR code selector initialized",
		zap.Int("qr_code_count", len(enabledQRCodes)),
		zap.String("polling_mode", pollingMode))

	return selector
}

// SelectQRCode 选择一个二维码
// @description 根据配置的轮询模式选择二维码
// @return *config.QRCode 选中的二维码
// @return error 选择错误
func (s *QRCodeSelector) SelectQRCode() (*config.QRCode, error) {
	if s == nil || len(s.qrCodes) == 0 {
		return nil, fmt.Errorf("no available QR codes")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var selected *config.QRCode

	switch s.pollingMode {
	case "round_robin":
		selected = s.selectRoundRobin()
	case "random":
		selected = s.selectRandom()
	case "least_used":
		selected = s.selectLeastUsed()
	default:
		selected = s.selectRoundRobin()
	}

	if selected == nil {
		return nil, fmt.Errorf("failed to select QR code")
	}

	// 更新使用统计
	s.usageCount[selected.ID]++
	s.lastUsedTime[selected.ID] = time.Now()

	logger.Debug("QR code selected",
		zap.String("qr_id", selected.ID),
		zap.String("mode", s.pollingMode),
		zap.Int("usage_count", s.usageCount[selected.ID]))

	return selected, nil
}

// selectRoundRobin 轮询选择
func (s *QRCodeSelector) selectRoundRobin() *config.QRCode {
	selected := &s.qrCodes[s.currentIndex]
	s.currentIndex = (s.currentIndex + 1) % len(s.qrCodes)
	return selected
}

// selectRandom 随机选择
func (s *QRCodeSelector) selectRandom() *config.QRCode {
	idx := rand.Intn(len(s.qrCodes))
	return &s.qrCodes[idx]
}

// selectLeastUsed 选择使用次数最少的
func (s *QRCodeSelector) selectLeastUsed() *config.QRCode {
	var selected *config.QRCode
	minUsage := -1

	for i := range s.qrCodes {
		qr := &s.qrCodes[i]
		usage := s.usageCount[qr.ID]

		if minUsage == -1 || usage < minUsage {
			minUsage = usage
			selected = qr
		}
	}

	return selected
}

// GetQRCodeByID 根据ID获取二维码
// @description 根据二维码ID获取二维码配置
// @param id 二维码ID
// @return *config.QRCode 二维码配置
// @return error 查询错误
func (s *QRCodeSelector) GetQRCodeByID(id string) (*config.QRCode, error) {
	if s == nil {
		return nil, fmt.Errorf("QR code selector not initialized")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.qrCodes {
		if s.qrCodes[i].ID == id {
			return &s.qrCodes[i], nil
		}
	}

	return nil, fmt.Errorf("QR code not found: %s", id)
}

// GetStats 获取使用统计
// @description 返回各个二维码的使用统计信息
// @return map[string]interface{} 统计信息
func (s *QRCodeSelector) GetStats() map[string]interface{} {
	if s == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make([]map[string]interface{}, 0, len(s.qrCodes))
	for _, qr := range s.qrCodes {
		stats = append(stats, map[string]interface{}{
			"id":             qr.ID,
			"usage_count":    s.usageCount[qr.ID],
			"last_used_time": s.lastUsedTime[qr.ID],
			"priority":       qr.Priority,
		})
	}

	return map[string]interface{}{
		"enabled":       true,
		"qr_code_count": len(s.qrCodes),
		"polling_mode":  s.pollingMode,
		"stats":         stats,
	}
}

// GetQRCodeCount 获取可用二维码数量
func (s *QRCodeSelector) GetQRCodeCount() int {
	if s == nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.qrCodes)
}

// IsEnabled 检查是否启用了多二维码模式
func (s *QRCodeSelector) IsEnabled() bool {
	return s != nil && len(s.qrCodes) > 0
}
