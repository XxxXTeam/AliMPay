package handler

import (
	"net/http"
	"time"

	"alimpay-go/internal/database"
	"alimpay-go/internal/model"
	"alimpay-go/internal/service"
	"alimpay-go/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminHandler 管理操作处理器
type AdminHandler struct {
	db         *database.DB
	codepay    *service.CodePayService
	merchantID string
}

// NewAdminHandler 创建管理处理器
func NewAdminHandler(db *database.DB, codepay *service.CodePayService) *AdminHandler {
	merchantInfo := codepay.GetMerchantInfo()
	return &AdminHandler{
		db:         db,
		codepay:    codepay,
		merchantID: merchantInfo["id"].(string),
	}
}

// HandleAdmin 处理管理操作（支持session和参数两种认证方式）
func (h *AdminHandler) HandleAdmin(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing action parameter",
		})
		return
	}

	switch action {
	case "pay", "mark_paid":
		h.handleMarkPaid(c)
	case "cancel":
		h.handleCancelOrder(c)
	case "refund":
		h.handleRefundOrder(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid action. Supported: pay, cancel, refund",
		})
	}
}

// HandleAdminAction 处理认证后的管理操作（基于session）
func (h *AdminHandler) HandleAdminAction(c *gin.Context) {
	// 从session获取商户ID
	merchantID, exists := c.Get("admin_merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Not authenticated",
		})
		return
	}

	// 解析请求
	var req struct {
		Action     string `json:"action" binding:"required"`
		TradeNo    string `json:"trade_no"`
		OutTradeNo string `json:"out_trade_no"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// 执行操作
	switch req.Action {
	case "pay", "mark_paid":
		h.markOrderPaid(c, merchantID.(string), req.TradeNo, req.OutTradeNo)
	case "cancel":
		h.cancelOrder(c, merchantID.(string), req.TradeNo)
	case "refund":
		h.refundOrder(c, merchantID.(string), req.TradeNo)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid action. Supported: pay, cancel, refund",
		})
	}
}

// HandleDashboard 渲染管理后台页面
func (h *AdminHandler) HandleDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_dashboard.html", nil)
}

// HandleGetOrders 获取订单列表（API）
func (h *AdminHandler) HandleGetOrders(c *gin.Context) {
	// 获取最近100个订单
	orders, err := h.db.GetOrders(h.codepay.GetMerchantID(), 100)
	if err != nil {
		logger.Error("Failed to get orders", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg":  "Failed to get orders",
		})
		return
	}

	// 转换为API格式
	var orderList []map[string]interface{}
	for _, order := range orders {
		orderList = append(orderList, map[string]interface{}{
			"trade_no":       order.ID,
			"out_trade_no":   order.OutTradeNo,
			"name":           order.Name,
			"price":          order.Price,
			"payment_amount": order.PaymentAmount,
			"status":         order.Status,
			"add_time":       order.AddTime,
			"pay_time":       order.PayTime,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   1,
		"msg":    "success",
		"orders": orderList,
	})
}

// handleMarkPaid 手动标记订单为已支付
func (h *AdminHandler) handleMarkPaid(c *gin.Context) {
	// 获取参数
	pid := c.Query("pid")
	key := c.Query("key")
	tradeNo := c.Query("trade_no")
	outTradeNo := c.Query("out_trade_no")

	// 验证必需参数
	if pid == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing required parameters: pid, key",
		})
		return
	}

	if tradeNo == "" && outTradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing order identifier: trade_no or out_trade_no required",
		})
		return
	}

	// 验证商户密钥
	merchantInfo := h.codepay.GetMerchantInfo()
	if pid != merchantInfo["id"].(string) || key != merchantInfo["key"].(string) {
		logger.Warn("Invalid admin credentials",
			zap.String("pid", pid),
			zap.String("ip", c.ClientIP()))
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid merchant credentials",
		})
		return
	}

	// 查询订单
	var order *model.Order
	var err error

	if tradeNo != "" {
		order, err = h.db.GetOrderByID(tradeNo)
	} else {
		order, err = h.db.GetOrderByOutTradeNo(outTradeNo, pid)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to query order: " + err.Error(),
		})
		return
	}

	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Order not found",
		})
		return
	}

	// 检查订单状态
	if order.Status == model.OrderStatusPaid {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Order already paid",
			"order": gin.H{
				"trade_no":     order.ID,
				"out_trade_no": order.OutTradeNo,
				"status":       "paid",
				"pay_time":     order.PayTime,
			},
		})
		return
	}

	// 更新订单状态为已支付
	payTime := time.Now()
	if err := h.db.UpdateOrderStatus(order.ID, model.OrderStatusPaid, payTime); err != nil {
		logger.Error("Failed to update order status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update order status: " + err.Error(),
		})
		return
	}

	logger.Info("Order manually marked as paid",
		zap.String("trade_no", order.ID),
		zap.String("out_trade_no", order.OutTradeNo),
		zap.String("operator_ip", c.ClientIP()))

	// 发送通知给商户
	notifySuccess := false
	var notifyError string

	if order.NotifyURL != "" {
		if err := h.codepay.SendNotification(order); err != nil {
			logger.Error("Failed to send notification",
				zap.String("trade_no", order.ID),
				zap.Error(err))
			notifyError = err.Error()
		} else {
			notifySuccess = true
		}
	}

	// 返回成功响应
	response := gin.H{
		"success": true,
		"message": "Order marked as paid successfully",
		"order": gin.H{
			"trade_no":       order.ID,
			"out_trade_no":   order.OutTradeNo,
			"status":         "paid",
			"pay_time":       payTime.Format("2006-01-02 15:04:05"),
			"payment_amount": order.PaymentAmount,
		},
	}

	if order.NotifyURL != "" {
		response["notification"] = gin.H{
			"sent":  notifySuccess,
			"url":   order.NotifyURL,
			"error": notifyError,
		}
	}

	c.JSON(http.StatusOK, response)
}

// handleCancelOrder 取消订单
func (h *AdminHandler) handleCancelOrder(c *gin.Context) {
	// 获取参数
	pid := c.Query("pid")
	key := c.Query("key")
	tradeNo := c.Query("trade_no")

	// 验证必需参数
	if pid == "" || key == "" || tradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing required parameters: pid, key, trade_no",
		})
		return
	}

	// 验证商户密钥
	merchantInfo := h.codepay.GetMerchantInfo()
	if pid != merchantInfo["id"].(string) || key != merchantInfo["key"].(string) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid merchant credentials",
		})
		return
	}

	// 查询订单
	order, err := h.db.GetOrderByID(tradeNo)
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Order not found",
		})
		return
	}

	// 更新订单状态为已关闭
	if err := h.db.UpdateOrderStatus(order.ID, model.OrderStatusClosed, time.Now()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to cancel order: " + err.Error(),
		})
		return
	}

	logger.Info("Order cancelled",
		zap.String("trade_no", order.ID),
		zap.String("operator_ip", c.ClientIP()))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order cancelled successfully",
		"order": gin.H{
			"trade_no":     order.ID,
			"out_trade_no": order.OutTradeNo,
			"status":       "closed",
		},
	})
}

// handleRefundOrder 退款订单
func (h *AdminHandler) handleRefundOrder(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"error":   "Refund function not implemented yet",
		"message": "Please process refunds manually through Alipay",
	})
}

// markOrderPaid 标记订单为已支付（基于session，简化版）
func (h *AdminHandler) markOrderPaid(c *gin.Context, merchantID, tradeNo, outTradeNo string) {
	// 查询订单
	var order *model.Order
	var err error

	if tradeNo != "" {
		order, err = h.db.GetOrderByID(tradeNo)
	} else if outTradeNo != "" {
		order, err = h.db.GetOrderByOutTradeNo(outTradeNo, merchantID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing order identifier: trade_no or out_trade_no required",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to query order: " + err.Error(),
		})
		return
	}

	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Order not found",
		})
		return
	}

	// 检查订单状态
	if order.Status == model.OrderStatusPaid {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Order already paid",
			"order": gin.H{
				"trade_no":     order.ID,
				"out_trade_no": order.OutTradeNo,
				"status":       "paid",
				"pay_time":     order.PayTime,
			},
		})
		return
	}

	// 更新订单状态为已支付
	payTime := time.Now()
	if err := h.db.UpdateOrderStatus(order.ID, model.OrderStatusPaid, payTime); err != nil {
		logger.Error("Failed to update order status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update order status: " + err.Error(),
		})
		return
	}

	logger.Info("Order manually marked as paid (session auth)",
		zap.String("trade_no", order.ID),
		zap.String("out_trade_no", order.OutTradeNo),
		zap.String("operator_ip", c.ClientIP()))

	// 发送通知给商户
	notifySuccess := false
	var notifyError string

	if order.NotifyURL != "" {
		if err := h.codepay.SendNotification(order); err != nil {
			logger.Error("Failed to send notification",
				zap.String("trade_no", order.ID),
				zap.Error(err))
			notifyError = err.Error()
		} else {
			notifySuccess = true
		}
	}

	// 返回成功响应
	response := gin.H{
		"success": true,
		"message": "Order marked as paid successfully",
		"order": gin.H{
			"trade_no":       order.ID,
			"out_trade_no":   order.OutTradeNo,
			"status":         "paid",
			"pay_time":       payTime.Format("2006-01-02 15:04:05"),
			"payment_amount": order.PaymentAmount,
		},
	}

	if order.NotifyURL != "" {
		response["notification"] = gin.H{
			"sent":  notifySuccess,
			"url":   order.NotifyURL,
			"error": notifyError,
		}
	}

	c.JSON(http.StatusOK, response)
}

// cancelOrder 取消订单（基于session，简化版）
func (h *AdminHandler) cancelOrder(c *gin.Context, merchantID, tradeNo string) {
	if tradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing required parameter: trade_no",
		})
		return
	}

	// 查询订单
	order, err := h.db.GetOrderByID(tradeNo)
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Order not found",
		})
		return
	}

	// 更新订单状态为已关闭
	if err := h.db.UpdateOrderStatus(order.ID, model.OrderStatusClosed, time.Now()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to cancel order: " + err.Error(),
		})
		return
	}

	logger.Info("Order cancelled (session auth)",
		zap.String("trade_no", order.ID),
		zap.String("operator_ip", c.ClientIP()))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order cancelled successfully",
		"order": gin.H{
			"trade_no":     order.ID,
			"out_trade_no": order.OutTradeNo,
			"status":       "closed",
		},
	})
}

// refundOrder 退款订单（基于session，简化版）
func (h *AdminHandler) refundOrder(c *gin.Context, merchantID, tradeNo string) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"error":   "Refund function not implemented yet",
		"message": "Please process refunds manually through Alipay",
	})
}
