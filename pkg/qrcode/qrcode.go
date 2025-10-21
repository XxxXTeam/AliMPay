package qrcode

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/skip2/go-qrcode"
)

// Generator 二维码生成器
type Generator struct {
	size   int
	margin int
}

// NewGenerator 创建二维码生成器
func NewGenerator(size, margin int) *Generator {
	if size <= 0 {
		size = 256
	}
	if margin < 0 {
		margin = 10
	}

	return &Generator{
		size:   size,
		margin: margin,
	}
}

// GenerateToBase64 生成base64编码的二维码
func (g *Generator) GenerateToBase64(content string) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to create QR code: %w", err)
	}

	qr.DisableBorder = g.margin == 0

	// 生成PNG格式的二维码
	pngData, err := qr.PNG(g.size)
	if err != nil {
		return "", fmt.Errorf("failed to generate PNG: %w", err)
	}

	// 转换为base64
	base64Str := base64.StdEncoding.EncodeToString(pngData)
	return base64Str, nil
}

// GenerateToFile 生成二维码并保存到文件
func (g *Generator) GenerateToFile(content, filePath string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 生成二维码
	err := qrcode.WriteFile(content, qrcode.Medium, g.size, filePath)
	if err != nil {
		return fmt.Errorf("failed to write QR code file: %w", err)
	}

	return nil
}

// GenerateToBytes 生成二维码字节数据
func (g *Generator) GenerateToBytes(content string) ([]byte, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	qr.DisableBorder = g.margin == 0

	pngData, err := qr.PNG(g.size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PNG: %w", err)
	}

	return pngData, nil
}

// GenerateFromURL 使用在线API生成二维码（备用方案）
func (g *Generator) GenerateFromURL(content string) ([]byte, error) {
	// 使用公共API生成二维码
	apiURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=%dx%d&data=%s",
		g.size, g.size, content)

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送请求
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to request QR code API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("QR code API returned status: %d", resp.StatusCode)
	}

	// 读取响应数据
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read QR code data: %w", err)
	}

	return data, nil
}

// GenerateURLFromAPI 生成在线二维码URL
func (g *Generator) GenerateURLFromAPI(content string) string {
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=%dx%d&data=%s",
		g.size, g.size, content)
}
