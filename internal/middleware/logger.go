package middleware

import (
	"time"

	"alimpay-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		latencyMs := float64(latency.Milliseconds())

		// 获取状态码
		statusCode := c.Writer.Status()

		// 忽略健康检查等不重要的日志
		if shouldSkipLog(path) {
			return
		}

		// 使用彩色日志
		method := c.Request.Method
		clientIP := c.ClientIP()

		// 根据状态码决定日志级别
		if statusCode >= 500 {
			logger.Error("Server Error",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", clientIP),
				zap.Int("status", statusCode),
				zap.Float64("latency_ms", latencyMs),
				zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			)
		} else if statusCode >= 400 {
			logger.Warn("Client Error",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", clientIP),
				zap.Int("status", statusCode),
				zap.Float64("latency_ms", latencyMs),
			)
		} else {
			// 成功请求使用精简日志
			logger.Request(method, path, clientIP, statusCode, latencyMs)
		}
	}
}

// shouldSkipLog 判断是否跳过日志记录
func shouldSkipLog(path string) bool {
	skipPaths := []string{
		"/health",
		"/ping",
		"/metrics",
		"/favicon.ico",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// Recovery 恢复中间件（带彩色日志）
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
