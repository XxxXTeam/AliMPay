package utils

import (
	"fmt"

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
