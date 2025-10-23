package utils

import (
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
)

// GetBaseURL 从请求中获取基础URL
// 如果配置了baseURL则直接使用，否则从请求中自动获取
func GetBaseURL(c *gin.Context, configBaseURL string) string {
	// 如果配置了基础URL，直接使用
	if configBaseURL != "" {
		return configBaseURL
	}

	// 自动从请求中获取
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	// 检查 X-Forwarded-Proto 头
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	// 获取Host
	host := c.Request.Host
	if host == "" {
		host = c.GetHeader("Host")
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

// GenerateAlipayDeepLink 生成支付宝直接拉起支付的深链接
// qrCodeID: 支付宝二维码ID
// amount: 支付金额
// remark: 备注信息（可选）
func GenerateAlipayDeepLink(qrCodeID string, amount float64, remark string) string {
	if qrCodeID == "" {
		return ""
	}

	// 构建支付宝二维码URL，格式: https://qr.alipay.com/{qrCodeId}?amount={amount}&remark={remark}
	alipayQRURL := fmt.Sprintf("https://qr.alipay.com/%s", qrCodeID)
	
	// 添加查询参数
	params := url.Values{}
	if amount > 0 {
		params.Add("amount", fmt.Sprintf("%.2f", amount))
	}
	if remark != "" {
		params.Add("remark", remark)
	}
	
	if len(params) > 0 {
		alipayQRURL += "?" + params.Encode()
	}

	// 构建深链接，使用 appId=20000056（转账到卡）
	deepLink := fmt.Sprintf("alipays://platformapi/startapp?appId=20000056&url=%s",
		url.QueryEscape(alipayQRURL))

	return deepLink
}
