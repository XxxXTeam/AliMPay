// Package service 订单监听任务实现
// @author AliMPay Team
// @description 提供订单监听任务的具体实现
package service

import (
	"context"
	"fmt"
	"time"

	"alimpay-go/internal/model"
	"alimpay-go/pkg/logger"

	"go.uber.org/zap"
)

// OrderMonitorTask 订单监听任务
// @description 监听单个订单的支付状态变化
type OrderMonitorTask struct {
	order   *model.Order
	monitor *MonitorService
}

// NewOrderMonitorTask 创建订单监听任务
// @description 为指定订单创建监听任务
// @param order 要监听的订单
// @param monitor 监听服务
// @return *OrderMonitorTask 任务实例
func NewOrderMonitorTask(order *model.Order, monitor *MonitorService) *OrderMonitorTask {
	return &OrderMonitorTask{
		order:   order,
		monitor: monitor,
	}
}

// Execute 执行订单监听任务
// @description 查询支付宝账单并尝试匹配订单
// @param ctx 上下文
// @return error 执行错误
func (t *OrderMonitorTask) Execute(ctx context.Context) error {
	// 检查订单当前状态
	currentOrder, err := t.monitor.db.GetOrderByID(t.order.ID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if currentOrder == nil || currentOrder.Status == model.OrderStatusPaid {
		return nil // 订单不存在或已支付
	}

	// 检查订单是否超时
	orderAge := time.Since(currentOrder.AddTime)
	if orderAge > 10*time.Minute {
		return nil // 超过10分钟不再监听
	}

	// 获取订单对应的账单查询服务
	billQueryService := t.monitor.GetBillQueryServiceForOrder(currentOrder)
	if billQueryService == nil {
		return nil // 账单查询服务不可用
	}

	// 查询支付宝账单（使用订单对应的API）
	var bills []BillRecord
	if currentOrder.QRCodeID != "" {
		// 如果订单有二维码ID，查询该二维码对应的账单
		bills, err = t.monitor.queryRecentBillsForQRCode(currentOrder.QRCodeID)
		if err != nil {
			logger.Debug("Failed to query bills for QR code, fallback to default",
				zap.String("qr_code_id", currentOrder.QRCodeID),
				zap.Error(err))
			// 如果失败，尝试使用默认服务
			bills, err = t.monitor.queryRecentBills()
			if err != nil {
				return err
			}
		}
	} else {
		// 使用默认账单查询
		bills, err = t.monitor.queryRecentBills()
		if err != nil {
			return err
		}
	}

	// 尝试匹配账单
	for _, bill := range bills {
		matched := false

		if t.monitor.cfg.Payment.BusinessQRMode.Enabled {
			matched = t.matchBusinessModeBill(bill)
		} else {
			matched = t.matchTraditionalModeBill(bill)
		}

		if matched {
			// 更新订单状态
			if err := t.monitor.updateOrderToPaid(currentOrder, bill.TradeNo); err != nil {
				logger.Error("Failed to update order status",
					zap.String("order_id", currentOrder.ID),
					zap.Error(err))
			}
			return nil
		}
	}

	return nil
}

// matchBusinessModeBill 匹配经营码模式账单
// @description 根据金额和时间匹配
// @param bill 账单记录
// @return bool 是否匹配
func (t *OrderMonitorTask) matchBusinessModeBill(bill BillRecord) bool {
	// 检查金额
	if fmt.Sprintf("%.2f", bill.Amount) != fmt.Sprintf("%.2f", t.order.PaymentAmount) {
		return false
	}

	// 解析支付时间
	billTime, err := time.ParseInLocation("2006-01-02 15:04:05", bill.TransDate, time.Local)
	if err != nil {
		return false
	}

	// 验证时间（支付必须在订单创建之后）
	timeDiff := billTime.Sub(t.order.AddTime)
	if timeDiff < 0 {
		return false
	}

	// 检查时间容差
	tolerance := time.Duration(t.monitor.cfg.Payment.BusinessQRMode.MatchTolerance) * time.Second
	return timeDiff <= tolerance
}

// matchTraditionalModeBill 匹配传统模式账单
// @description 根据备注（订单号）和金额匹配
// @param bill 账单记录
// @return bool 是否匹配
func (t *OrderMonitorTask) matchTraditionalModeBill(bill BillRecord) bool {
	// 检查备注是否为订单号
	if bill.Remark != t.order.OutTradeNo {
		return false
	}

	// 验证金额
	return fmt.Sprintf("%.2f", bill.Amount) == fmt.Sprintf("%.2f", t.order.Price)
}
