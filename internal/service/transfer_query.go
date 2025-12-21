package service

import (
	"encoding/json"
	"fmt"
	"time"

	"alimpay-go/internal/pkg/logger"

	"go.uber.org/zap"
)

// TransferQueryService 转账查询服务（备用支付检测方案）
type TransferQueryService struct {
	client *AlipayClient
}

// NewTransferQueryService 创建转账查询服务
func NewTransferQueryService(client *AlipayClient) *TransferQueryService {
	return &TransferQueryService{
		client: client,
	}
}

// TransferQueryRequest 转账查询请求
type TransferQueryRequest struct {
	OutBizNo string `json:"out_biz_no"` // 商户订单号
	OrderID  string `json:"order_id"`   // 支付宝订单号
	PayDate  string `json:"pay_date"`   // 支付时间
}

// TransferQueryResponse 转账查询响应
type TransferQueryResponse struct {
	Code           string `json:"code"`
	Msg            string `json:"msg"`
	SubCode        string `json:"sub_code"`
	SubMsg         string `json:"sub_msg"`
	OrderID        string `json:"order_id"`
	Status         string `json:"status"`
	PayDate        string `json:"pay_date"`
	ArrivalTimeEnd string `json:"arrival_time_end"`
	OrderFee       string `json:"order_fee"`
	ErrorCode      string `json:"error_code"`
	FailReason     string `json:"fail_reason"`
}

// QueryTransferOrder 查询转账订单
func (s *TransferQueryService) QueryTransferOrder(outBizNo string) (*TransferQueryResponse, error) {
	logger.Info("Querying transfer order", zap.String("out_biz_no", outBizNo))

	// 构建业务参数
	bizContent := map[string]interface{}{
		"out_biz_no":   outBizNo,
		"product_code": "TRANS_ACCOUNT_NO_PWD", // 单笔转账到支付宝账户
	}
	bizContentJSON, _ := json.Marshal(bizContent)

	// 构建请求参数
	params := s.client.buildRequestParams("alipay.fund.trans.order.query", string(bizContentJSON))

	// 生成签名
	sign, err := s.client.generateSign(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sign: %w", err)
	}
	params["sign"] = sign

	// 发送请求
	resp, err := s.client.doRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	// 解析响应
	var response struct {
		AlipayFundTransOrderQueryResponse TransferQueryResponse `json:"alipay_fund_trans_order_query_response"`
		Sign                              string                `json:"sign"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.AlipayFundTransOrderQueryResponse.Code != "10000" {
		logger.Error("Transfer query API error",
			zap.String("code", response.AlipayFundTransOrderQueryResponse.Code),
			zap.String("msg", response.AlipayFundTransOrderQueryResponse.Msg),
			zap.String("sub_code", response.AlipayFundTransOrderQueryResponse.SubCode),
			zap.String("sub_msg", response.AlipayFundTransOrderQueryResponse.SubMsg))
		return nil, fmt.Errorf("transfer query error: %s - %s",
			response.AlipayFundTransOrderQueryResponse.Code,
			response.AlipayFundTransOrderQueryResponse.Msg)
	}

	logger.Info("Transfer query successful",
		zap.String("order_id", response.AlipayFundTransOrderQueryResponse.OrderID),
		zap.String("status", response.AlipayFundTransOrderQueryResponse.Status))

	return &response.AlipayFundTransOrderQueryResponse, nil
}

// CheckTransferSuccess 检查转账是否成功
func (s *TransferQueryService) CheckTransferSuccess(outBizNo string) (bool, error) {
	result, err := s.QueryTransferOrder(outBizNo)
	if err != nil {
		return false, err
	}

	// SUCCESS: 转账成功
	// FAIL: 转账失败
	// DEALING: 处理中
	// REFUND: 退票
	return result.Status == "SUCCESS", nil
}

// MonitorTransferOrder 监控转账订单（轮询直到完成或超时）
func (s *TransferQueryService) MonitorTransferOrder(outBizNo string, timeout time.Duration) (bool, error) {
	startTime := time.Now()
	ticker := time.NewTicker(3 * time.Second) // 每3秒查询一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			success, err := s.CheckTransferSuccess(outBizNo)
			if err != nil {
				// 如果是订单不存在的错误，继续等待
				logger.Debug("Transfer query error, will retry", zap.Error(err))
				continue
			}

			if success {
				logger.Info("Transfer completed successfully", zap.String("out_biz_no", outBizNo))
				return true, nil
			}

			// 检查是否超时
			if time.Since(startTime) > timeout {
				logger.Warn("Transfer monitoring timeout",
					zap.String("out_biz_no", outBizNo),
					zap.Duration("timeout", timeout))
				return false, fmt.Errorf("monitoring timeout after %v", timeout)
			}

		case <-time.After(timeout):
			return false, fmt.Errorf("monitoring timeout")
		}
	}
}
