package utils

import (
	"strings"
	"testing"
)

func TestGenerateAlipayDeepLink(t *testing.T) {
	tests := []struct {
		name      string
		qrCodeID  string
		amount    float64
		remark    string
		wantEmpty bool
		contains  []string
	}{
		{
			name:      "Empty QR code ID should return empty string",
			qrCodeID:  "",
			amount:    1.0,
			remark:    "test",
			wantEmpty: true,
		},
		{
			name:     "Valid QR code ID with amount and remark",
			qrCodeID: "fkx12345678901234",
			amount:   1.23,
			remark:   "测试订单",
			contains: []string{
				"alipays://platformapi/startapp",
				"appId=20000056",
				"url=",
				"https%3A%2F%2Fqr.alipay.com%2Ffkx12345678901234",
				"amount%3D1.23",
				"remark%3D",
			},
		},
		{
			name:     "QR code ID with amount only",
			qrCodeID: "fkx98765432109876",
			amount:   99.99,
			remark:   "",
			contains: []string{
				"alipays://platformapi/startapp",
				"appId=20000056",
				"https%3A%2F%2Fqr.alipay.com%2Ffkx98765432109876",
				"amount%3D99.99",
			},
		},
		{
			name:     "QR code ID without amount or remark",
			qrCodeID: "fkx11111111111111",
			amount:   0,
			remark:   "",
			contains: []string{
				"alipays://platformapi/startapp",
				"appId=20000056",
				"https%3A%2F%2Fqr.alipay.com%2Ffkx11111111111111",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateAlipayDeepLink(tt.qrCodeID, tt.amount, tt.remark)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("GenerateAlipayDeepLink() = %v, want empty string", got)
				}
				return
			}

			if got == "" {
				t.Errorf("GenerateAlipayDeepLink() returned empty string, want non-empty")
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("GenerateAlipayDeepLink() = %v, should contain %v", got, want)
				}
			}
		})
	}
}

func TestGenerateAlipayDeepLinkFormat(t *testing.T) {
	qrCodeID := "fkx12345678901234"
	amount := 1.50
	remark := "test order"

	deepLink := GenerateAlipayDeepLink(qrCodeID, amount, remark)

	// Check that the link starts with the correct scheme
	if !strings.HasPrefix(deepLink, "alipays://platformapi/startapp?") {
		t.Errorf("Deep link should start with 'alipays://platformapi/startapp?', got: %s", deepLink)
	}

	// Check that appId is present
	if !strings.Contains(deepLink, "appId=20000056") {
		t.Errorf("Deep link should contain 'appId=20000056', got: %s", deepLink)
	}

	// Check that url parameter is present and properly encoded
	if !strings.Contains(deepLink, "url=https%3A%2F%2Fqr.alipay.com%2F") {
		t.Errorf("Deep link should contain encoded qr.alipay.com URL, got: %s", deepLink)
	}
}
