package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidateOrderParams 验证订单参数
func ValidateOrderParams(params map[string]string) error {
	// 必需字段
	required := []string{"pid", "type", "out_trade_no", "name", "money"}
	for _, field := range required {
		if params[field] == "" {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// 验证 PID（商户ID）
	if err := ValidatePID(params["pid"]); err != nil {
		return err
	}

	// 验证订单号
	if err := ValidateOutTradeNo(params["out_trade_no"]); err != nil {
		return err
	}

	// 验证金额
	if err := ValidateMoney(params["money"]); err != nil {
		return err
	}

	// 验证支付类型
	if err := ValidatePaymentType(params["type"]); err != nil {
		return err
	}

	// 验证URL（如果提供）
	if params["notify_url"] != "" {
		if err := ValidateURL(params["notify_url"]); err != nil {
			return fmt.Errorf("invalid notify_url: %w", err)
		}
	}

	if params["return_url"] != "" {
		if err := ValidateURL(params["return_url"]); err != nil {
			return fmt.Errorf("invalid return_url: %w", err)
		}
	}

	return nil
}

// ValidatePID 验证商户ID
func ValidatePID(pid string) error {
	if len(pid) == 0 || len(pid) > 32 {
		return fmt.Errorf("invalid pid length")
	}

	// 只允许数字和字母
	matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", pid)
	if !matched {
		return fmt.Errorf("pid contains invalid characters")
	}

	return nil
}

// ValidateOutTradeNo 验证订单号
func ValidateOutTradeNo(outTradeNo string) error {
	if len(outTradeNo) == 0 || len(outTradeNo) > 64 {
		return fmt.Errorf("invalid out_trade_no length")
	}

	// 只允许数字、字母、下划线和连字符
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", outTradeNo)
	if !matched {
		return fmt.Errorf("out_trade_no contains invalid characters")
	}

	return nil
}

// ValidateMoney 验证金额
func ValidateMoney(money string) error {
	// 验证金额格式（允许负数用于格式检测，但后续会拒绝）
	matched, _ := regexp.MatchString(`^-?\d+(\.\d{1,2})?$`, money)
	if !matched {
		return fmt.Errorf("invalid money format")
	}

	// 转换并严格验证金额
	amount, err := strconv.ParseFloat(money, 64)
	if err != nil {
		return fmt.Errorf("invalid money value")
	}

	if amount <= 0 {
		return fmt.Errorf("money must be greater than 0 (0 yuan purchase not allowed)")
	}

	if amount < 0.01 {
		return fmt.Errorf("money must be at least 0.01 yuan")
	}

	if amount > 99999.99 {
		return fmt.Errorf("money exceeds maximum limit (99999.99)")
	}

	return nil
}

// ValidatePaymentType 验证支付类型
func ValidatePaymentType(paymentType string) error {
	validTypes := map[string]bool{
		"alipay": true,
		"wxpay":  true,
	}

	if !validTypes[paymentType] {
		return fmt.Errorf("unsupported payment type: %s", paymentType)
	}

	return nil
}

// ValidateURL 验证URL格式
func ValidateURL(urlStr string) error {
	if len(urlStr) > 500 {
		return fmt.Errorf("url too long")
	}

	// 简单的URL验证
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return fmt.Errorf("url must start with http:// or https://")
	}

	return nil
}

// SanitizeString 清理字符串，防止XSS
func SanitizeString(s string) string {
	// 移除潜在的危险字符
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	s = strings.ReplaceAll(s, "&", "&amp;")
	return s
}

// ValidateSignType 验证签名类型
func ValidateSignType(signType string) error {
	validTypes := map[string]bool{
		"MD5":  true,
		"RSA":  true,
		"RSA2": true,
	}

	if !validTypes[signType] {
		return fmt.Errorf("unsupported sign type: %s", signType)
	}

	return nil
}
