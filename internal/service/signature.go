package service

import (
	"github.com/alimpay/alimpay-go/pkg/logger"
	"github.com/alimpay/alimpay-go/pkg/utils"
	"go.uber.org/zap"
)

// ValidateSignature 验证请求签名
func (s *CodePayService) ValidateSignature(params map[string]string) bool {
	receivedSign := params["sign"]
	if receivedSign == "" {
		logger.Warn("Missing signature in request")
		return false
	}

	// 计算签名
	calculatedSign := utils.GenerateSign(params, s.merchantKey)

	// 对比签名
	if receivedSign != calculatedSign {
		logger.Warn("Signature mismatch",
			zap.String("received", receivedSign),
			zap.String("calculated", calculatedSign))
		return false
	}

	return true
}
