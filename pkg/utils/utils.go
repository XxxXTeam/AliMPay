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
	if _, err := rand.Read(b); err != nil {
		// 如果随机数生成失败，使用时间戳作为fallback
		return fmt.Sprintf("%032x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// RandomInt 生成随机整数
func RandomInt(min, max int) int {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// 如果随机数生成失败，使用时间戳作为fallback
		return min + int(time.Now().UnixNano()%(int64(max-min+1)))
	}
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

/*
 * GenerateSign 生成签名（兼容易支付标准）
 * @description 按照易支付/码支付标准生成MD5签名
 * @param params map[string]string 参数Map
 * @param key string 商户密钥
 * @return string 32位小写MD5签名
 *
 * 签名算法：
 * 1. 过滤空值参数和 sign、sign_type
 * 2. 按参数名ASCII码升序排序
 * 3. 使用URL键值对格式拼接成字符串（key1=value1&key2=value2）
 * 4. 在字符串末尾拼接商户密钥
 * 5. MD5加密并转小写
 */
func GenerateSign(params map[string]string, key string) string {
	// 1. 移除空值和签名相关参数
	filtered := make(map[string]string)
	for k, v := range params {
		if v != "" && k != "sign" && k != "sign_type" {
			filtered[k] = v
		}
	}

	// 2. 按键名ASCII码排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 3. 构建签名字符串 key1=value1&key2=value2
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, filtered[k]))
	}

	// 4. 拼接商户密钥
	signStr := strings.Join(parts, "&")
	signStrWithKey := signStr + key

	// 5. MD5加密（小写）
	return strings.ToLower(MD5(signStrWithKey))
}

/*
 * VerifySign 验证签名（兼容易支付标准）
 * @description 验证请求签名是否正确，支持大小写不敏感比对
 * @param params map[string]string 请求参数Map
 * @param key string 商户密钥
 * @return bool 签名是否正确
 */
func VerifySign(params map[string]string, key string) bool {
	receivedSign := params["sign"]
	if receivedSign == "" {
		return false
	}

	// 生成期望的签名
	expectedSign := GenerateSign(params, key)

	// 大小写不敏感比对（易支付兼容性）
	return strings.ToLower(receivedSign) == strings.ToLower(expectedSign)
}

/*
 * VerifySignDebug 验证签名（调试版本）
 * @description 验证签名并返回详细的调试信息
 * @param params map[string]string 请求参数Map
 * @param key string 商户密钥
 * @return bool 签名是否正确
 * @return string 调试信息
 */
func VerifySignDebug(params map[string]string, key string) (bool, string) {
	receivedSign := params["sign"]
	if receivedSign == "" {
		return false, "签名参数为空"
	}

	// 构建签名字符串用于调试
	filtered := make(map[string]string)
	for k, v := range params {
		if v != "" && k != "sign" && k != "sign_type" {
			filtered[k] = v
		}
	}

	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, filtered[k]))
	}

	signStr := strings.Join(parts, "&")
	signStrWithKey := signStr + key
	expectedSign := strings.ToLower(MD5(signStrWithKey))

	debugInfo := fmt.Sprintf(
		"签名验证详情:\n"+
			"  参与签名的参数: %v\n"+
			"  签名字符串: %s\n"+
			"  加上密钥后: %s\n"+
			"  计算出的签名: %s\n"+
			"  接收到的签名: %s\n"+
			"  验证结果: %v",
		filtered,
		signStr,
		signStrWithKey,
		expectedSign,
		receivedSign,
		strings.ToLower(receivedSign) == strings.ToLower(expectedSign),
	)

	return strings.ToLower(receivedSign) == strings.ToLower(expectedSign), debugInfo
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
