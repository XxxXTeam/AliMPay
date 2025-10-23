/*
Package middleware 缓存控制中间件
Author: AliMPay Team
Description: 为静态资源和页面添加适当的缓存控制头

功能:
  - 静态资源长期缓存
  - HTML页面短期缓存
  - ETag支持
  - Cache-Control配置
*/
package middleware

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

/*
CacheControl 缓存控制中间件配置
类型:
  - NoCache: 不缓存，每次都重新请求
  - ShortCache: 短期缓存（5分钟）
  - LongCache: 长期缓存（1年）
*/
type CacheType int

const (
	NoCache CacheType = iota
	ShortCache
	LongCache
)

/*
CacheMiddleware 返回缓存控制中间件
参数:
  - cacheType: 缓存类型

使用示例:

	router.Use(middleware.CacheMiddleware(middleware.LongCache))
*/
func CacheMiddleware(cacheType CacheType) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch cacheType {
		case NoCache:
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		case ShortCache:
			c.Header("Cache-Control", "public, max-age=300") // 5分钟
		case LongCache:
			c.Header("Cache-Control", "public, max-age=31536000, immutable") // 1年
		}
		c.Next()
	}
}

/*
StaticCacheMiddleware 静态资源缓存中间件
功能:
  - 根据文件扩展名设置不同的缓存策略
  - CSS/JS: 长期缓存（1年）
  - 图片: 长期缓存（1年）
  - 字体: 长期缓存（1年）
  - HTML: 短期缓存（5分钟）
*/
func StaticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 判断文件类型
		switch {
		case strings.HasSuffix(path, ".css"):
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			c.Header("Content-Type", "text/css; charset=utf-8")
		case strings.HasSuffix(path, ".js"):
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			c.Header("Content-Type", "application/javascript; charset=utf-8")
		case strings.HasSuffix(path, ".png"):
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			c.Header("Content-Type", "image/png")
		case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			c.Header("Content-Type", "image/jpeg")
		case strings.HasSuffix(path, ".gif"):
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			c.Header("Content-Type", "image/gif")
		case strings.HasSuffix(path, ".svg"):
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			c.Header("Content-Type", "image/svg+xml")
		case strings.HasSuffix(path, ".woff"), strings.HasSuffix(path, ".woff2"):
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		case strings.HasSuffix(path, ".html"):
			c.Header("Cache-Control", "public, max-age=300") // 5分钟
		default:
			c.Header("Cache-Control", "public, max-age=3600") // 1小时
		}

		c.Next()
	}
}

/*
ETagMiddleware ETag缓存中间件
功能:
  - 为响应生成ETag
  - 支持If-None-Match条件请求
  - 返回304 Not Modified（如果内容未变化）

使用示例:

	router.Use(middleware.ETagMiddleware())
*/
func ETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求的If-None-Match头
		ifNoneMatch := c.GetHeader("If-None-Match")

		// 创建响应写入器包装
		writer := &etagResponseWriter{
			ResponseWriter: c.Writer,
			ifNoneMatch:    ifNoneMatch,
		}
		c.Writer = writer

		c.Next()

		// 如果有响应体，生成ETag
		if len(writer.body) > 0 {
			etag := generateETag(writer.body)
			c.Header("ETag", etag)

			// 如果ETag匹配，返回304
			if ifNoneMatch == etag {
				c.Status(304)
				c.Writer = writer.ResponseWriter // 恢复原始writer
				c.Abort()
				return
			}
		}

		// 写入实际响应
		writer.ResponseWriter.Write(writer.body)
	}
}

// etagResponseWriter ETag响应写入器
type etagResponseWriter struct {
	gin.ResponseWriter
	body        []byte
	ifNoneMatch string
}

// Write 实现io.Writer接口
func (w *etagResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return len(data), nil
}

// generateETag 生成ETag
func generateETag(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf(`"%x"`, hash)
}

/*
VersionedStaticMiddleware 带版本号的静态资源中间件
功能:
  - 添加版本号到Cache-Control
  - 支持查询参数版本控制
  - 强制缓存破坏

使用示例:

	router.Static("/static", "./static")
	router.Use(middleware.VersionedStaticMiddleware("v1.0.0"))
*/
func VersionedStaticMiddleware(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否有版本参数
		v := c.Query("v")
		if v != "" && v != version {
			// 版本不匹配，强制重新加载
			c.Header("Cache-Control", "no-cache, must-revalidate")
		} else {
			// 版本匹配或无版本，长期缓存
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			c.Header("X-Content-Version", version)
		}
		c.Next()
	}
}

/*
CompressMiddleware 压缩中间件（简化版）
功能:
  - 添加压缩相关头
  - 建议使用Nginx等反向代理进行实际压缩

使用示例:

	router.Use(middleware.CompressMiddleware())
*/
func CompressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置Vary头，告诉缓存服务器根据Accept-Encoding区分
		c.Header("Vary", "Accept-Encoding")
		c.Next()
	}
}

/*
LastModifiedMiddleware 最后修改时间中间件
功能:
  - 添加Last-Modified头
  - 支持If-Modified-Since条件请求

参数:
  - modTime: 资源最后修改时间
*/
func LastModifiedMiddleware(modTime time.Time) gin.HandlerFunc {
	lastModified := modTime.UTC().Format(time.RFC1123)
	return func(c *gin.Context) {
		// 设置Last-Modified头
		c.Header("Last-Modified", lastModified)

		// 检查If-Modified-Since
		ifModifiedSince := c.GetHeader("If-Modified-Since")
		if ifModifiedSince != "" {
			t, err := time.Parse(time.RFC1123, ifModifiedSince)
			if err == nil && !modTime.After(t) {
				c.Status(304)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
