package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"alimpay-go/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestIDKey 请求ID在上下文中的键名
const RequestIDKey = "request_id"

// generateRequestID 生成唯一请求ID
func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405.000")
	}
	return hex.EncodeToString(b)
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 生成并设置请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)

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
				zap.String("request_id", requestID),
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
				zap.String("request_id", requestID),
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
				requestID, _ := c.Get(RequestIDKey)
				logger.Error("Panic recovered",
					zap.Any("request_id", requestID),
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

// GetRequestID 从上下文获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
