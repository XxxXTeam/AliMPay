package model

import (
	"time"
)

// Order 订单模型
type Order struct {
	ID            string     `db:"id" json:"id"`
	OutTradeNo    string     `db:"out_trade_no" json:"out_trade_no"`
	Type          string     `db:"type" json:"type"`
	PID           string     `db:"pid" json:"pid"`
	Name          string     `db:"name" json:"name"`
	Price         float64    `db:"price" json:"price"`
	PaymentAmount float64    `db:"payment_amount" json:"payment_amount"`
	Status        int        `db:"status" json:"status"`
	AddTime       time.Time  `db:"add_time" json:"add_time"`
	PayTime       *time.Time `db:"pay_time" json:"pay_time,omitempty"`
	NotifyURL     string     `db:"notify_url" json:"notify_url"`
	ReturnURL     string     `db:"return_url" json:"return_url"`
	Sitename      string     `db:"sitename" json:"sitename"`
}

// OrderStatus 订单状态
const (
	OrderStatusPending = 0 // 待支付
	OrderStatusPaid    = 1 // 已支付
	OrderStatusClosed  = 2 // 已关闭
	OrderStatusRefund  = 3 // 已退款
)

// PaymentType 支付类型
const (
	PaymentTypeAlipay = "alipay"
)
