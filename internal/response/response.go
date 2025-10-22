package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一API响应结构
type Response struct {
	Code    int         `json:"code"`              // 状态码：1=成功，-1=失败
	Message string      `json:"msg,omitempty"`     // 消息
	Data    interface{} `json:"data,omitempty"`    // 数据
	Error   string      `json:"error,omitempty"`   // 错误信息
	Success bool        `json:"success,omitempty"` // 成功标志（用于管理接口）
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    1,
		Message: "SUCCESS",
		Data:    data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    1,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    -1,
		Message: message,
	})
}

// ErrorWithCode 带状态码的错误响应
func ErrorWithCode(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    -1,
		Message: message,
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    -1,
		Message: message,
		Data:    data,
	})
}

// AdminSuccess 管理接口成功响应
func AdminSuccess(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// AdminError 管理接口错误响应
func AdminError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"error":   message,
	})
}

// YiPaySuccess 易支付标准成功响应
func YiPaySuccess(c *gin.Context, data map[string]interface{}) {
	response := gin.H{
		"code": 1,
		"msg":  "SUCCESS",
	}
	for k, v := range data {
		response[k] = v
	}
	c.JSON(http.StatusOK, response)
}

// YiPayError 易支付标准错误响应
func YiPayError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"code": -1,
		"msg":  message,
	})
}

// Redirect 重定向响应
func Redirect(c *gin.Context, url string) {
	c.Redirect(http.StatusFound, url)
}

// HTML 渲染HTML响应
func HTML(c *gin.Context, template string, data interface{}) {
	c.HTML(http.StatusOK, template, data)
}

// JSON 通用JSON响应
func JSON(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// String 纯文本响应
func String(c *gin.Context, text string) {
	c.String(http.StatusOK, text)
}

// NotFound 404响应
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		Code:    -1,
		Message: "Resource not found",
	})
}

// Unauthorized 401响应
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    -1,
		Message: message,
	})
}

// Forbidden 403响应
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    -1,
		Message: message,
	})
}

// BadRequest 400响应
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    -1,
		Message: message,
	})
}

// InternalServerError 500响应
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    -1,
		Message: message,
	})
}
