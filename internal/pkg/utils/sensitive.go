package utils

import (
	"strings"
)

// MaskString 脱敏字符串（保留前后各n个字符）
func MaskString(s string, prefixLen, suffixLen int) string {
	if s == "" {
		return ""
	}

	length := len(s)
	if length <= prefixLen+suffixLen {
		return strings.Repeat("*", length)
	}

	prefix := s[:prefixLen]
	suffix := s[length-suffixLen:]
	middle := strings.Repeat("*", length-prefixLen-suffixLen)

	return prefix + middle + suffix
}

// MaskKey 脱敏密钥（显示前4位后4位）
func MaskKey(key string) string {
	return MaskString(key, 4, 4)
}

// MaskSign 脱敏签名（只显示前8位）
func MaskSign(sign string) string {
	if len(sign) <= 8 {
		return sign
	}
	return sign[:8] + "..."
}

// MaskPhone 脱敏手机号
func MaskPhone(phone string) string {
	return MaskString(phone, 3, 4)
}

// MaskEmail 脱敏邮箱
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return username + "@" + domain
	}

	maskedUsername := username[:1] + strings.Repeat("*", len(username)-2) + username[len(username)-1:]
	return maskedUsername + "@" + domain
}

// MaskOrderNo 脱敏订单号（保留前6位后4位）
func MaskOrderNo(orderNo string) string {
	return MaskString(orderNo, 6, 4)
}

// SanitizeResponse 清理响应中的敏感信息
func SanitizeResponse(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		switch key {
		case "key", "merchant_key", "app_key":
			// 密钥完全隐藏
			result[key] = "***"
		case "sign":
			// 签名脱敏
			if strVal, ok := value.(string); ok {
				result[key] = MaskSign(strVal)
			} else {
				result[key] = value
			}
		case "private_key", "alipay_public_key":
			// 私钥/公钥完全隐藏
			result[key] = "***HIDDEN***"
		default:
			result[key] = value
		}
	}

	return result
}
