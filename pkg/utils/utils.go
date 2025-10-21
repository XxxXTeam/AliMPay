package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

// GenerateTradeNo 生成交易号
func GenerateTradeNo() string {
	return fmt.Sprintf("%s%06d", time.Now().Format("20060102150405"), RandomInt(1, 999999))
}

// GenerateMerchantID 生成商户ID
func GenerateMerchantID() string {
	return fmt.Sprintf("1001%012d", RandomInt(0, 999999999999))
}

// GenerateMerchantKey 生成商户密钥
func GenerateMerchantKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// RandomInt 生成随机整数
func RandomInt(min, max int) int {
	b := make([]byte, 4)
	rand.Read(b)
	n := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if n < 0 {
		n = -n
	}
	return min + n%(max-min+1)
}

// MD5 计算MD5哈希
func MD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateSign 生成签名
func GenerateSign(params map[string]string, key string) string {
	// 移除空值
	filtered := make(map[string]string)
	for k, v := range params {
		if v != "" && k != "sign" && k != "sign_type" {
			filtered[k] = v
		}
	}

	// 按键名排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, filtered[k]))
	}

	signStr := strings.Join(parts, "&")
	return MD5(signStr + key)
}

// VerifySign 验证签名
func VerifySign(params map[string]string, key string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}

	expectedSign := GenerateSign(params, key)
	return sign == expectedSign
}

// FormatAmount 格式化金额（保留2位小数）
func FormatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// ParseTime 解析时间字符串
func ParseTime(timeStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
	}

	var lastErr error
	for _, layout := range layouts {
		t, err := time.Parse(layout, timeStr)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}

	return time.Time{}, lastErr
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// IsExpired 检查是否过期
func IsExpired(createTime time.Time, timeout int) bool {
	return time.Since(createTime) > time.Duration(timeout)*time.Second
}
