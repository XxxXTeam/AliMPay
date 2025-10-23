/*
Package middleware 路径规范化中间件
Author: AliMPay Team
Description: 处理URL路径中的异常情况

功能:
  - 去除路径中的多余斜杠
  - 规范化URL路径
  - 兼容各种客户端实现

示例:
  //submit -> /submit
  ///api -> /api
  /submit/ -> /submit
*/
package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// 预编译正则表达式
var multipleSlashes = regexp.MustCompile(`/{2,}`)

/*
PathNormalizer 路径规范化中间件
功能:
  - 将多个连续斜杠替换为单个斜杠
  - 去除路径末尾的斜杠（根路径除外）
  - 自动处理URL编码问题

使用示例:
  router.Use(middleware.PathNormalizer())

处理示例:
  //submit -> /submit
  ///api/query -> /api/query
  /submit/ -> /submit
  /api//order -> /api/order
*/
func PathNormalizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		originalPath := path

		// 1. 替换多个连续斜杠为单个斜杠
		path = multipleSlashes.ReplaceAllString(path, "/")

		// 2. 去除末尾斜杠（但保留根路径 "/"）
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			path = strings.TrimSuffix(path, "/")
		}

		// 3. 如果路径被修改了，进行内部重定向
		if path != originalPath {
			// 更新请求路径
			c.Request.URL.Path = path
			
			// 记录路径规范化（仅在debug模式）
			if gin.Mode() == gin.DebugMode {
				c.Set("original_path", originalPath)
				c.Set("normalized_path", path)
			}
		}

		c.Next()
	}
}

/*
StrictPathNormalizer 严格路径规范化中间件
功能: 除了规范化路径外，还会返回301重定向

使用场景:
  - SEO优化
  - 强制规范URL格式

注意:
  - 会导致额外的HTTP请求
  - 不适用于API接口
*/
func StrictPathNormalizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		originalPath := path

		// 规范化路径
		path = multipleSlashes.ReplaceAllString(path, "/")
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			path = strings.TrimSuffix(path, "/")
		}

		// 如果路径改变了，发送301重定向
		if path != originalPath {
			// 保留查询参数
			if c.Request.URL.RawQuery != "" {
				path = path + "?" + c.Request.URL.RawQuery
			}
			c.Redirect(http.StatusMovedPermanently, path)
			c.Abort()
			return
		}

		c.Next()
	}
}

/*
RemoveTrailingSlash 移除末尾斜杠中间件
功能: 仅移除路径末尾的斜杠

使用场景:
  - 简单的URL规范化
  - 不需要处理多余斜杠的场景
*/
func RemoveTrailingSlash() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			c.Request.URL.Path = strings.TrimSuffix(path, "/")
		}
		c.Next()
	}
}

