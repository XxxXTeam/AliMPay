// Package web 提供Web资源管理
// @author AliMPay Team
// @description 统一管理所有Web资源，包括HTML模板和静态文件
package web

import (
	"embed"
	"html/template"
	"io/fs"
)

// Templates 嵌入所有HTML模板文件
// @description 使用embed将模板文件嵌入到二进制文件中，便于部署
//
//go:embed templates/*.html
var Templates embed.FS

// Static 嵌入所有静态资源文件
// @description 使用embed将CSS、JS等静态文件嵌入到二进制文件中
//
//go:embed static/css/*.css static/js/*.js
var Static embed.FS

// ParseTemplates 解析所有模板文件
// @description 从embed.FS中解析HTML模板
// @return *template.Template 解析后的模板集合
// @return error 解析错误
func ParseTemplates() (*template.Template, error) {
	return template.ParseFS(Templates, "templates/*.html")
}

// GetTemplatesFS 获取模板文件系统
// @description 返回一个只包含templates目录的文件系统，用于Gin加载模板
// @return fs.FS 模板文件系统
// @return error 错误信息
func GetTemplatesFS() (fs.FS, error) {
	return fs.Sub(Templates, "templates")
}

// GetStaticFS 获取静态文件系统
// @description 返回一个只包含static目录的文件系统，用于Gin提供静态文件服务
// @return fs.FS 静态文件系统
// @return error 错误信息
func GetStaticFS() (fs.FS, error) {
	return fs.Sub(Static, "static")
}
