package service

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/alimpay/alimpay-go/internal/config"
	"github.com/alimpay/alimpay-go/pkg/logger"
	"go.uber.org/zap"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AlipayClient 支付宝客户端
type AlipayClient struct {
	cfg        *config.AlipayConfig
	httpClient *http.Client
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// BillQueryRequest 账单查询请求
type BillQueryRequest struct {
	StartTime string `json:"start_time"` // 开始时间 YYYY-MM-DD HH:mm:ss
	EndTime   string `json:"end_time"`   // 结束时间 YYYY-MM-DD HH:mm:ss
	PageNo    int    `json:"page_no"`    // 页码
	PageSize  int    `json:"page_size"`  // 每页大小
}

// BillQueryResponse 账单查询响应
type BillQueryResponse struct {
	Code       string       `json:"code"`
	Msg        string       `json:"msg"`
	SubCode    string       `json:"sub_code"`
	SubMsg     string       `json:"sub_msg"`
	DetailList []BillDetail `json:"detail_list"`
	PageNo     string       `json:"page_no"`
	PageSize   string       `json:"page_size"`
	TotalSize  string       `json:"total_size"`
}

// BillDetail 账单明细
type BillDetail struct {
	AccountLogID    string `json:"account_log_id"`  // 账务流水号
	AlipayOrderNo   string `json:"alipay_order_no"` // 支付宝交易号
	MerchantOrderNo string `json:"merchant_out_no"` // 商户订单号
	TransAmount     string `json:"trans_amount"`    // 交易金额
	TransMemo       string `json:"trans_memo"`      // 交易备注
	TransDt         string `json:"trans_dt"`        // 交易时间
	Direction       string `json:"direction"`       // 收入/支出
	OtherAccount    string `json:"other_account"`   // 对方账户
	Balance         string `json:"balance"`         // 账户余额
	Type            string `json:"type"`            // 业务类型
}

// NewAlipayClient 创建支付宝客户端
func NewAlipayClient(cfg *config.AlipayConfig) (*AlipayClient, error) {
	client := &AlipayClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// 解析私钥
	if err := client.parsePrivateKey(); err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// 解析公钥
	if err := client.parsePublicKey(); err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	logger.Info("Alipay client initialized successfully")
	return client, nil
}

// parsePrivateKey 解析应用私钥
func (c *AlipayClient) parsePrivateKey() error {
	privateKeyStr := c.cfg.PrivateKey

	// 如果私钥不包含 PEM 头尾，添加它们
	if !strings.Contains(privateKeyStr, "BEGIN") {
		privateKeyStr = fmt.Sprintf("-----BEGIN RSA PRIVATE KEY-----\n%s\n-----END RSA PRIVATE KEY-----", privateKeyStr)
	}

	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试 PKCS8 格式
		pkcs8Key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return fmt.Errorf("failed to parse private key: %v, %v", err, err2)
		}
		var ok bool
		privateKey, ok = pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("not RSA private key")
		}
	}

	c.privateKey = privateKey
	return nil
}

// parsePublicKey 解析支付宝公钥
func (c *AlipayClient) parsePublicKey() error {
	publicKeyStr := c.cfg.AlipayPublicKey

	// 如果公钥不包含 PEM 头尾，添加它们
	if !strings.Contains(publicKeyStr, "BEGIN") {
		publicKeyStr = fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", publicKeyStr)
	}

	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not RSA public key")
	}

	c.publicKey = publicKey
	return nil
}

// Sign 签名
func (c *AlipayClient) Sign(data string) (string, error) {
	hash := crypto.SHA256.New()
	hash.Write([]byte(data))
	hashed := hash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// Verify 验证签名
func (c *AlipayClient) Verify(data, sign string) error {
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return err
	}

	hash := crypto.SHA256.New()
	hash.Write([]byte(data))
	hashed := hash.Sum(nil)

	return rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, hashed, signBytes)
}

// buildRequestParams 构建请求参数
func (c *AlipayClient) buildRequestParams(method string, bizContent string) map[string]string {
	params := map[string]string{
		"app_id":      c.cfg.AppID,
		"method":      method,
		"format":      c.cfg.Format,
		"charset":     c.cfg.Charset,
		"sign_type":   c.cfg.SignType,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": bizContent,
	}

	return params
}

// generateSign 生成签名字符串
func (c *AlipayClient) generateSign(params map[string]string) (string, error) {
	// 排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接签名字符串
	var signStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
	}

	// 签名
	return c.Sign(signStr.String())
}

// 注意：本系统不使用 alipay.trade.query 接口（需要额外权限）
// 完全依赖账单查询接口，和PHP版本保持一致

// QueryBills 查询账单
func (c *AlipayClient) QueryBills(startTime, endTime string, pageNo, pageSize int) (*BillQueryResponse, error) {
	logger.Info("Querying Alipay bills",
		zap.String("start_time", startTime),
		zap.String("end_time", endTime),
		zap.Int("page_no", pageNo),
		zap.Int("page_size", pageSize),
		zap.String("app_id", c.cfg.AppID[:min(len(c.cfg.AppID), 10)]+"...")) // 只显示前10位

	// 构建业务参数
	bizContent := map[string]interface{}{
		"start_time": startTime,
		"end_time":   endTime,
		"page_no":    pageNo,
		"page_size":  pageSize,
	}
	bizContentJSON, _ := json.Marshal(bizContent)

	// 构建请求参数
	params := c.buildRequestParams("alipay.data.bill.accountlog.query", string(bizContentJSON))

	// 生成签名
	sign, err := c.generateSign(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sign: %w", err)
	}
	params["sign"] = sign

	// 发送请求
	resp, err := c.doRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	// 解析响应
	var response struct {
		AlipayDataBillAccountlogQueryResponse BillQueryResponse `json:"alipay_data_bill_accountlog_query_response"`
		Sign                                  string            `json:"sign"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 验证响应签名（生产环境应该启用）
	// TODO: 实现响应签名验证

	if response.AlipayDataBillAccountlogQueryResponse.Code != "10000" {
		logger.Error("Alipay API error",
			zap.String("code", response.AlipayDataBillAccountlogQueryResponse.Code),
			zap.String("msg", response.AlipayDataBillAccountlogQueryResponse.Msg),
			zap.String("sub_code", response.AlipayDataBillAccountlogQueryResponse.SubCode),
			zap.String("sub_msg", response.AlipayDataBillAccountlogQueryResponse.SubMsg))
		return nil, fmt.Errorf("alipay API error: %s - %s",
			response.AlipayDataBillAccountlogQueryResponse.Code,
			response.AlipayDataBillAccountlogQueryResponse.Msg)
	}

	logger.Info("Bills query successful",
		zap.Int("count", len(response.AlipayDataBillAccountlogQueryResponse.DetailList)))

	return &response.AlipayDataBillAccountlogQueryResponse, nil
}

// doRequest 发送HTTP请求
func (c *AlipayClient) doRequest(params map[string]string) ([]byte, error) {
	// 构建请求URL
	reqURL := c.cfg.ServerURL

	// 构建表单数据
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	logger.Debug("Sending request to Alipay",
		zap.String("url", reqURL),
		zap.String("method", params["method"]))

	// 发送请求
	resp, err := c.httpClient.Post(reqURL, "application/x-www-form-urlencoded;charset=utf-8", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 详细日志：记录完整响应内容（用于诊断API问题）
	logger.Info("==================== 支付宝API完整响应 ====================")
	logger.Info("API响应状态",
		zap.Int("status_code", resp.StatusCode),
		zap.String("content_type", resp.Header.Get("Content-Type")))
	logger.Info("完整响应内容", zap.String("response", string(body)))
	logger.Info("==========================================================")

	return body, nil
}

// Validate 验证配置
func (c *AlipayClient) Validate() error {
	if c.cfg.AppID == "" {
		return fmt.Errorf("app_id is required")
	}
	if c.cfg.PrivateKey == "" {
		return fmt.Errorf("private_key is required")
	}
	if c.cfg.AlipayPublicKey == "" {
		return fmt.Errorf("alipay_public_key is required")
	}
	if c.cfg.ServerURL == "" {
		return fmt.Errorf("server_url is required")
	}
	return nil
}
