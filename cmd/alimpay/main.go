// Package main 应用程序入口
// @author AliMPay Team
// @description AliMPay 支付系统主程序，负责初始化各个模块并启动HTTP服务
package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alimpay-go/internal/config"
	"alimpay-go/internal/database"
	"alimpay-go/internal/handler"
	"alimpay-go/internal/middleware"
	"alimpay-go/internal/service"
	"alimpay-go/pkg/logger"
	"alimpay-go/web"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 设置全局时区为北京时间（和PHP版本保持一致）
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Printf("Failed to load timezone: %v\n", err)
		os.Exit(1)
	}
	time.Local = loc

	// 解析命令行参数
	configPath := flag.String("config", "./configs/config.yaml", "Path to configuration file")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	logCfg := &logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		Output:     cfg.Logging.Output,
		FilePath:   cfg.Logging.FilePath,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
	}

	if err := logger.Init(logCfg); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 美化的启动信息
	logger.Highlight("AliMPay Golang Version Starting",
		zap.String("version", "1.0.0"),
		zap.String("config", *configPath),
		zap.String("timezone", "Asia/Shanghai"))

	// 初始化数据库
	dbCfg := &database.Config{
		Type:            cfg.Database.Type,
		Path:            cfg.Database.Path,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	db, err := database.Init(dbCfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// 初始化服务
	codepayService, err := service.NewCodePayService(cfg, db)
	if err != nil {
		logger.Fatal("Failed to initialize CodePay service", zap.Error(err))
	}

	monitorService, err := service.NewMonitorService(cfg, db, codepayService)
	if err != nil {
		logger.Fatal("Failed to initialize Monitor service", zap.Error(err))
	}

	// 启动监控服务
	if err := monitorService.Start(); err != nil {
		logger.Fatal("Failed to start monitor service", zap.Error(err))
	}
	defer monitorService.Stop()

	// 启动自动回调服务
	autoCallback := service.NewAutoCallbackService(db, codepayService)
	autoCallback.Start()
	defer autoCallback.Stop()

	// 初始化HTTP服务器
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 使用自定义中间件（彩色日志）
	router := gin.New()
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())

	// 从嵌入的文件系统加载HTML模板
	tmpl := template.Must(template.New("").ParseFS(web.Templates, "templates/*.html"))
	router.SetHTMLTemplate(tmpl)

	logger.Success("Templates loaded from embedded filesystem", zap.Int("count", len(tmpl.Templates())))

	// 静态文件 - 使用嵌入的文件系统
	staticFS, err := web.GetStaticFS()
	if err != nil {
		logger.Fatal("Failed to get static filesystem", zap.Error(err))
	}
	router.StaticFS("/static", http.FS(staticFS))

	// 初始化handlers
	apiHandler := handler.NewAPIHandler(codepayService, monitorService, cfg)
	submitHandler := handler.NewSubmitHandler(codepayService, cfg)
	healthHandler := handler.NewHealthHandler(db, codepayService, monitorService)
	qrcodeHandler := handler.NewQRCodeHandler(cfg)
	adminHandler := handler.NewAdminHandler(db, codepayService)
	yipayHandler := handler.NewYiPayHandler(db, codepayService, cfg)
	payHandler := handler.NewPayHandler(db, cfg)
	wsHandler := handler.NewWebSocketHandler(db)

	// 注册路由 - 易支付/码支付标准接口

	// API接口（兼容模式）
	router.GET("/api", apiHandler.HandleAction)
	router.POST("/api", apiHandler.HandleAction)

	// MAPI接口（码支付标准）
	router.GET("/mapi", yipayHandler.HandleMAPI)
	router.POST("/mapi", yipayHandler.HandleMAPI)

	// Submit接口（创建支付）
	router.GET("/submit", submitHandler.HandleSubmit)
	router.POST("/submit", submitHandler.HandleSubmit)
	router.GET("/submit.php", submitHandler.HandleSubmit)
	router.POST("/submit.php", submitHandler.HandleSubmit)

	// API提交接口（易支付标准）
	router.GET("/api/submit", yipayHandler.HandleSubmitAPI)
	router.POST("/api/submit", yipayHandler.HandleSubmitAPI)

	// 查询接口
	router.GET("/api/query", yipayHandler.HandleQueryMerchant)
	router.POST("/api/query", yipayHandler.HandleQueryMerchant)
	router.GET("/api/order", yipayHandler.HandleQueryOrder)
	router.POST("/api/order", yipayHandler.HandleQueryOrder)

	// 订单管理
	router.GET("/api/close", yipayHandler.HandleClose)
	router.POST("/api/close", yipayHandler.HandleClose)
	router.GET("/api/refund", yipayHandler.HandleRefund)
	router.POST("/api/refund", yipayHandler.HandleRefund)

	// 回调接口
	router.GET("/notify", yipayHandler.HandleCallback)
	router.POST("/notify", yipayHandler.HandleCallback)
	router.GET("/notify.php", yipayHandler.HandleCallback)
	router.POST("/notify.php", yipayHandler.HandleCallback)
	router.GET("/callback", yipayHandler.HandleCallback)
	router.POST("/callback", yipayHandler.HandleCallback)

	// 签名验证接口
	router.GET("/api/checksign", yipayHandler.HandleCheckSign)
	router.POST("/api/checksign", yipayHandler.HandleCheckSign)

	// 系统接口
	router.GET("/health", healthHandler.HandleHealth)
	router.GET("/qrcode", qrcodeHandler.HandleQRCode)
	router.GET("/pay", payHandler.HandlePayPage) // 支付页面（扫码后跳转）

	// WebSocket接口 - 实时订单状态推送
	router.GET("/ws/order", wsHandler.HandleWebSocket)

	// 管理接口
	router.GET("/admin/dashboard", adminHandler.HandleDashboard) // 管理后台页面
	router.GET("/admin/orders", adminHandler.HandleGetOrders)    // 获取订单列表
	router.GET("/admin", adminHandler.HandleAdmin)               // 管理操作API
	router.POST("/admin", adminHandler.HandleAdmin)

	// 启动HTTP服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// 优雅退出
	go func() {
		if err := router.Run(addr); err != nil {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// 打印商户信息（美化版）
	merchantInfo := codepayService.GetMerchantInfo()

	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║         🚀 AliMPay Golang Version Started            ║")
	fmt.Println("╠════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Server Address:  http://%-28s ║\n", addr)
	fmt.Printf("║  Merchant ID:     %-35s ║\n", merchantInfo["id"])
	fmt.Printf("║  Merchant Key:    %-35s ║\n", merchantInfo["key"])
	fmt.Printf("║  Monitor:         %-35s ║\n", fmt.Sprintf("Enabled (Interval: %ds)", cfg.Monitor.Interval))
	fmt.Printf("║  Mode:            %-35s ║\n", cfg.Server.Mode)
	fmt.Println("╚════════════════════════════════════════════════════════╝\n")

	logger.Success("Server started successfully",
		zap.String("address", addr),
		zap.String("merchant_id", merchantInfo["id"].(string)))

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println()
	logger.Warn("Received shutdown signal, gracefully stopping...")
	logger.Sync()
}
