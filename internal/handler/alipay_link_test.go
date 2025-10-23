package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"alimpay-go/internal/config"

	"github.com/gin-gonic/gin"
)

func TestAlipayLinkHandler_HandleGenerateLink(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		qrCodeIDConfig string
		queryParams    string
		wantStatus     int
		wantCode       int
		checkFields    []string
	}{
		{
			name:           "Valid request with config qr_code_id",
			qrCodeIDConfig: "fkx12345678901234",
			queryParams:    "?amount=1.23&remark=测试",
			wantStatus:     http.StatusOK,
			wantCode:       1,
			checkFields:    []string{"alipay_deep_link", "qr_code_id", "amount"},
		},
		{
			name:           "Valid request with custom qr_code_id",
			qrCodeIDConfig: "",
			queryParams:    "?qr_code_id=custom123&amount=5.00",
			wantStatus:     http.StatusOK,
			wantCode:       1,
			checkFields:    []string{"alipay_deep_link"},
		},
		{
			name:           "Missing qr_code_id",
			qrCodeIDConfig: "",
			queryParams:    "?amount=1.00",
			wantStatus:     http.StatusBadRequest,
			wantCode:       -1,
		},
		{
			name:           "Invalid amount format",
			qrCodeIDConfig: "fkx12345678901234",
			queryParams:    "?amount=invalid",
			wantStatus:     http.StatusBadRequest,
			wantCode:       -1,
		},
		{
			name:           "Negative amount",
			qrCodeIDConfig: "fkx12345678901234",
			queryParams:    "?amount=-1.00",
			wantStatus:     http.StatusBadRequest,
			wantCode:       -1,
		},
		{
			name:           "Amount exceeds limit",
			qrCodeIDConfig: "fkx12345678901234",
			queryParams:    "?amount=100000.00",
			wantStatus:     http.StatusBadRequest,
			wantCode:       -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			cfg := &config.Config{
				Payment: config.PaymentConfig{
					BusinessQRMode: config.BusinessQRMode{
						QRCodeID: tt.qrCodeIDConfig,
					},
				},
			}

			// Create handler
			handler := NewAlipayLinkHandler(cfg)

			// Create test request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/alipay/link"+tt.queryParams, nil)

			// Call handler
			handler.HandleGenerateLink(c)

			// Check status code
			if w.Code != tt.wantStatus {
				t.Errorf("HandleGenerateLink() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Check code
			if code, ok := response["code"].(float64); ok {
				if int(code) != tt.wantCode {
					t.Errorf("HandleGenerateLink() code = %v, want %v", int(code), tt.wantCode)
				}
			}

			// Check required fields for successful requests
			if tt.wantStatus == http.StatusOK {
				for _, field := range tt.checkFields {
					if _, ok := response[field]; !ok {
						t.Errorf("HandleGenerateLink() response missing field: %v", field)
					}
				}

				// Verify deep link format
				if deepLink, ok := response["alipay_deep_link"].(string); ok {
					if deepLink == "" {
						t.Error("HandleGenerateLink() alipay_deep_link is empty")
					}
					if len(deepLink) < 50 {
						t.Errorf("HandleGenerateLink() alipay_deep_link seems too short: %v", deepLink)
					}
				}
			}
		})
	}
}
