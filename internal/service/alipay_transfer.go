package service

import (
	"fmt"
	"net/url"
	"strings"

	"alimpay-go/internal/config"
	"alimpay-go/pkg/logger"

	"go.uber.org/zap"
)

// AlipayTransfer 支付宝转账服务
type AlipayTransfer struct {
	cfg *config.AlipayConfig
}

// NewAlipayTransfer 创建支付宝转账服务
func NewAlipayTransfer(cfg *config.AlipayConfig) *AlipayTransfer {
	return &AlipayTransfer{
		cfg: cfg,
	}
}

// GenerateTransferURL 生成转账URL
func (at *AlipayTransfer) GenerateTransferURL(amount float64, memo, userID string) string {
	// 如果未指定userID，使用配置中的默认值
	if userID == "" {
		userID = at.cfg.TransferUserID
	}

	// 获取防风控配置
	antiRiskCfg := config.Get().Payment.AntiRiskURL

	if antiRiskCfg.Enabled {
		return at.generateAntiRiskURL(amount, memo, userID, &antiRiskCfg)
	}

	return at.generateSimpleURL(amount, memo, userID)
}

// generateSimpleURL 生成简单转账URL
func (at *AlipayTransfer) generateSimpleURL(amount float64, memo, userID string) string {
	params := url.Values{}
	params.Set("appId", "09999988")
	params.Set("actionType", "toAccount")
	params.Set("goBack", "NO")
	params.Set("amount", fmt.Sprintf("%.2f", amount))
	params.Set("userId", userID)
	params.Set("memo", memo)

	transferURL := fmt.Sprintf("alipays://platformapi/startapp?%s", params.Encode())

	logger.Info("Generated simple transfer URL",
		zap.Float64("amount", amount),
		zap.String("memo", memo),
		zap.String("user_id", userID))

	return transferURL
}

// generateAntiRiskURL 生成防风控转账URL（多层嵌套）
func (at *AlipayTransfer) generateAntiRiskURL(amount float64, memo, userID string, cfg *config.AntiRiskURLConfig) string {
	// 第1层：最内层转账URL
	innerParams := url.Values{}
	innerParams.Set("appId", cfg.InnerAppID)
	innerParams.Set("actionType", "toAccount")
	innerParams.Set("goBack", "NO")
	innerParams.Set("amount", fmt.Sprintf("%.2f", amount))
	innerParams.Set("userId", userID)
	innerParams.Set("memo", memo)

	innerURL := fmt.Sprintf("alipays://platformapi/startapp?%s", innerParams.Encode())

	// 第2层：scheme包装
	layer2URL := fmt.Sprintf("%s?scheme=%s", cfg.RenderSchemeURL, url.QueryEscape(innerURL))

	// 第3层：外层app包装
	layer3Params := url.Values{}
	layer3Params.Set("appId", cfg.OuterAppID)
	layer3Params.Set("url", layer2URL)

	layer3URL := fmt.Sprintf("alipays://platformapi/startapp?%s", layer3Params.Encode())

	// 第4层：再次scheme包装
	layer4URL := fmt.Sprintf("%s?scheme=%s", cfg.RenderSchemeURL, url.QueryEscape(layer3URL))

	// 第5层：最外层mdeduct包装
	finalURL := fmt.Sprintf("%s?scheme=%s", cfg.MdeductLandingURL, url.QueryEscape(layer4URL))

	logger.Info("Generated anti-risk transfer URL",
		zap.Float64("amount", amount),
		zap.String("memo", memo),
		zap.String("user_id", userID),
		zap.String("outer_app_id", cfg.OuterAppID),
		zap.String("inner_app_id", cfg.InnerAppID))

	return finalURL
}

// ParseAntiRiskURL 解析防风控URL（用于验证）
func (at *AlipayTransfer) ParseAntiRiskURL(transferURL string) map[string]string {
	result := make(map[string]string)
	antiRiskCfg := config.Get().Payment.AntiRiskURL

	// 检查最外层
	if !strings.HasPrefix(transferURL, antiRiskCfg.MdeductLandingURL) {
		result["valid"] = "false"
		result["error"] = "Invalid outer layer URL"
		return result
	}

	result["layer1"] = "valid"

	// 逐层解析
	currentURL := transferURL
	for i := 1; i <= 5; i++ {
		schemeIndex := strings.Index(currentURL, "scheme=")
		if schemeIndex == -1 {
			break
		}

		encodedURL := currentURL[schemeIndex+7:]
		decodedURL, err := url.QueryUnescape(encodedURL)
		if err != nil {
			result["error"] = fmt.Sprintf("Failed to decode layer %d", i)
			return result
		}

		result[fmt.Sprintf("layer%d", i+1)] = "valid"
		currentURL = decodedURL
	}

	result["valid"] = "true"
	return result
}
