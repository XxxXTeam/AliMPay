package service

import (
	"fmt"
	"time"

	"github.com/alimpay/alimpay-go/internal/config"
	"github.com/alimpay/alimpay-go/pkg/logger"
	"go.uber.org/zap"
)

// BillQueryService 账单查询服务
type BillQueryService struct {
	alipayClient *AlipayClient
}

// NewBillQueryService 创建账单查询服务
func NewBillQueryService(cfg *config.AlipayConfig) (*BillQueryService, error) {
	alipayClient, err := NewAlipayClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create alipay client: %w", err)
	}

	// 验证配置
	if err := alipayClient.Validate(); err != nil {
		return nil, fmt.Errorf("invalid alipay config: %w", err)
	}

	return &BillQueryService{
		alipayClient: alipayClient,
	}, nil
}

// QueryBills 查询账单
func (s *BillQueryService) QueryBills(startTime, endTime string, pageNo, pageSize int) (map[string]interface{}, error) {
	// 设置默认值
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 || pageSize > 2000 {
		pageSize = 2000
	}

	// 验证时间格式
	if err := s.validateTimeFormat(startTime); err != nil {
		return nil, fmt.Errorf("invalid start_time: %w", err)
	}
	if err := s.validateTimeFormat(endTime); err != nil {
		return nil, fmt.Errorf("invalid end_time: %w", err)
	}

	logger.Info("Querying bills",
		zap.String("start_time", startTime),
		zap.String("end_time", endTime),
		zap.Int("page_no", pageNo),
		zap.Int("page_size", pageSize))

	// 调用支付宝API
	resp, err := s.alipayClient.QueryBills(startTime, endTime, pageNo, pageSize)
	if err != nil {
		return nil, err
	}

	// 格式化返回结果
	result := map[string]interface{}{
		"success":   true,
		"data":      s.formatBillData(resp),
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}

	return result, nil
}

// QueryTodayBills 查询今日账单
func (s *BillQueryService) QueryTodayBills() (map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	startTime := today + " 00:00:00"
	endTime := today + " 23:59:59"

	return s.QueryBills(startTime, endTime, 1, 2000)
}

// QueryYesterdayBills 查询昨日账单
func (s *BillQueryService) QueryYesterdayBills() (map[string]interface{}, error) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	startTime := yesterday + " 00:00:00"
	endTime := yesterday + " 23:59:59"

	return s.QueryBills(startTime, endTime, 1, 2000)
}

// QueryBillsByDate 查询指定日期账单
func (s *BillQueryService) QueryBillsByDate(date string) (map[string]interface{}, error) {
	startTime := date + " 00:00:00"
	endTime := date + " 23:59:59"

	return s.QueryBills(startTime, endTime, 1, 2000)
}

// QueryRecentBills 查询最近N小时的账单
func (s *BillQueryService) QueryRecentBills(hoursBack int) (map[string]interface{}, error) {
	// 减去5分钟以避免时间同步问题
	endTime := time.Now().Add(-5 * time.Minute).Format("2006-01-02 15:04:05")
	startTime := time.Now().Add(-time.Duration(hoursBack)*time.Hour - 5*time.Minute).Format("2006-01-02 15:04:05")

	logger.Info("Querying recent bills",
		zap.String("start_time", startTime),
		zap.String("end_time", endTime),
		zap.Int("hours_back", hoursBack))

	return s.QueryBills(startTime, endTime, 1, 100)
}

// QueryBillsInTimeRange 查询指定时间范围的账单
func (s *BillQueryService) QueryBillsInTimeRange(startTime, endTime string) (map[string]interface{}, error) {
	return s.QueryBills(startTime, endTime, 1, 100)
}

// validateTimeFormat 验证时间格式
func (s *BillQueryService) validateTimeFormat(timeStr string) error {
	_, err := time.Parse("2006-01-02 15:04:05", timeStr)
	return err
}

// formatBillData 格式化账单数据
func (s *BillQueryService) formatBillData(resp *BillQueryResponse) map[string]interface{} {
	// 转换详细列表
	detailList := make([]map[string]interface{}, 0, len(resp.DetailList))
	for _, detail := range resp.DetailList {
		detailList = append(detailList, map[string]interface{}{
			"account_log_id":  detail.AccountLogID,
			"alipay_order_no": detail.AlipayOrderNo,
			"merchant_out_no": detail.MerchantOrderNo,
			"trans_amount":    detail.TransAmount,
			"trans_memo":      detail.TransMemo,
			"trans_dt":        detail.TransDt,
			"direction":       detail.Direction,
			"other_account":   detail.OtherAccount,
			"balance":         detail.Balance,
			"type":            detail.Type,
		})
	}

	return map[string]interface{}{
		"detail_list": detailList,
		"page_no":     resp.PageNo,
		"page_size":   resp.PageSize,
		"total_size":  resp.TotalSize,
	}
}

// FindPaymentByMemo 根据备注查找支付记录
func (s *BillQueryService) FindPaymentByMemo(billData map[string]interface{}, orderNo string, expectedAmount float64) map[string]interface{} {
	// 提取账单列表
	detailList, ok := billData["detail_list"].([]map[string]interface{})
	if !ok {
		logger.Warn("Invalid bill data structure")
		return nil
	}

	for _, bill := range detailList {
		// 只处理收入类型
		direction, _ := bill["direction"].(string)
		if direction != "收入" {
			continue
		}

		// 获取备注和金额
		memo, _ := bill["trans_memo"].(string)
		amountStr, _ := bill["trans_amount"].(string)

		// 匹配订单号
		if memo != orderNo {
			continue
		}

		// 解析金额
		var amount float64
		fmt.Sscanf(amountStr, "%f", &amount)

		// 匹配金额（允许0.01的误差）
		if amount < expectedAmount-0.01 || amount > expectedAmount+0.01 {
			logger.Debug("Order matched but amount mismatch",
				zap.String("order_no", orderNo),
				zap.Float64("expected", expectedAmount),
				zap.Float64("actual", amount))
			continue
		}

		// 找到匹配的支付记录
		logger.Info("Payment match found",
			zap.String("order_no", orderNo),
			zap.Float64("amount", amount))

		return bill
	}

	return nil
}
