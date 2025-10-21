package handler

import (
	"net/http"
	"time"

	"github.com/alimpay/alimpay-go/internal/database"
	"github.com/alimpay/alimpay-go/internal/model"
	"github.com/alimpay/alimpay-go/internal/service"
	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	db      *database.DB
	codepay *service.CodePayService
	monitor *service.MonitorService
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(db *database.DB, codepay *service.CodePayService, monitor *service.MonitorService) *HealthHandler {
	return &HealthHandler{
		db:      db,
		codepay: codepay,
		monitor: monitor,
	}
}

// HandleHealth 处理健康检查请求
func (h *HealthHandler) HandleHealth(c *gin.Context) {
	action := c.Query("action")
	if action == "" {
		action = "status"
	}

	switch action {
	case "status", "":
		h.handleStatus(c)
	case "monitor", "trigger_monitor", "run_monitor":
		h.handleMonitor(c)
	case "cleanup":
		h.handleCleanup(c)
	case "debug":
		h.handleDebug(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid action. Supported: status, monitor, cleanup, debug",
		})
	}
}

// handleStatus 处理状态查询
func (h *HealthHandler) handleStatus(c *gin.Context) {
	// 统计订单数量
	totalOrders, _ := h.db.CountOrders(nil)
	pendingStatus := model.OrderStatusPending
	unpaidOrders, _ := h.db.CountOrders(&pendingStatus)

	// 获取监控状态
	monitorStatus := h.monitor.GetStatus()

	// 构建响应
	response := gin.H{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"system":    "AliMPay Golang Version",
		"status":    "ok",
		"services": gin.H{
			"database": gin.H{
				"status":        "healthy",
				"total_orders":  totalOrders,
				"unpaid_orders": unpaidOrders,
			},
			"monitoring": monitorStatus,
		},
		"counters": gin.H{
			"total_orders":  totalOrders,
			"unpaid_orders": unpaidOrders,
			"paid_orders":   totalOrders - unpaidOrders,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// handleMonitor 手动触发监控
func (h *HealthHandler) handleMonitor(c *gin.Context) {
	go h.monitor.RunMonitoringCycle()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Monitoring cycle triggered",
		"data": gin.H{
			"action":    "monitor_triggered",
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// handleCleanup 清理过期订单
func (h *HealthHandler) handleCleanup(c *gin.Context) {
	count, err := h.codepay.CleanupExpiredOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cleanup completed",
		"data": gin.H{
			"deleted_count": count,
			"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// handleDebug 调试信息
func (h *HealthHandler) handleDebug(c *gin.Context) {
	// 获取最近的订单（使用数据库提供的方法）
	recentOrders, err := h.db.GetRecentOrders(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"recent_orders": recentOrders,
			"monitor":       h.monitor.GetStatus(),
			"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}
