// Package main 应用程序入口
// @author AliMPay Team
// @description AliMPay 支付系统主程序，负责初始化各个模块并启动HTTP服务
package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	router.Use(middleware.PathNormalizer()) // 路径规范化，处理//submit等情况

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
	adminWsHandler := handler.NewAdminWebSocketHandler(db)

	// 初始化管理员认证中间件
	merchantInfo := codepayService.GetMerchantInfo()
	adminAuth := middleware.NewAdminAuthMiddleware(
		merchantInfo["id"].(string),
		merchantInfo["key"].(string),
	)

	// 注册路由 - 易支付/码支付标准接口

	// API接口（兼容模式） - 支持.php后缀
	router.GET("/api", apiHandler.HandleAction)
	router.POST("/api", apiHandler.HandleAction)
	router.GET("/api.php", apiHandler.HandleAction)
	router.POST("/api.php", apiHandler.HandleAction)

	// MAPI接口（码支付标准） - 支持.php后缀
	router.GET("/mapi", yipayHandler.HandleMAPI)
	router.POST("/mapi", yipayHandler.HandleMAPI)
	router.GET("/mapi.php", yipayHandler.HandleMAPI)
	router.POST("/mapi.php", yipayHandler.HandleMAPI)

	// Submit接口（创建支付） - 支持.php后缀
	router.GET("/submit", submitHandler.HandleSubmit)
	router.POST("/submit", submitHandler.HandleSubmit)
	router.GET("/submit.php", submitHandler.HandleSubmit)
	router.POST("/submit.php", submitHandler.HandleSubmit)

	// API提交接口（易支付标准） - 支持.php后缀
	router.GET("/api/submit", yipayHandler.HandleSubmitAPI)
	router.POST("/api/submit", yipayHandler.HandleSubmitAPI)
	router.GET("/api/submit.php", yipayHandler.HandleSubmitAPI)
	router.POST("/api/submit.php", yipayHandler.HandleSubmitAPI)

	// 查询接口 - 支持.php后缀
	router.GET("/api/query", yipayHandler.HandleQueryMerchant)
	router.POST("/api/query", yipayHandler.HandleQueryMerchant)
	router.GET("/api/query.php", yipayHandler.HandleQueryMerchant)
	router.POST("/api/query.php", yipayHandler.HandleQueryMerchant)
	router.GET("/api/order", yipayHandler.HandleQueryOrder)
	router.POST("/api/order", yipayHandler.HandleQueryOrder)
	router.GET("/api/order.php", yipayHandler.HandleQueryOrder)
	router.POST("/api/order.php", yipayHandler.HandleQueryOrder)

	// 订单管理 - 支持.php后缀
	router.GET("/api/close", yipayHandler.HandleClose)
	router.POST("/api/close", yipayHandler.HandleClose)
	router.GET("/api/close.php", yipayHandler.HandleClose)
	router.POST("/api/close.php", yipayHandler.HandleClose)
	router.GET("/api/refund", yipayHandler.HandleRefund)
	router.POST("/api/refund", yipayHandler.HandleRefund)
	router.GET("/api/refund.php", yipayHandler.HandleRefund)
	router.POST("/api/refund.php", yipayHandler.HandleRefund)

	// 回调接口 - 支持.php后缀
	router.GET("/notify", yipayHandler.HandleCallback)
	router.POST("/notify", yipayHandler.HandleCallback)
	router.GET("/notify.php", yipayHandler.HandleCallback)
	router.POST("/notify.php", yipayHandler.HandleCallback)
	router.GET("/callback", yipayHandler.HandleCallback)
	router.POST("/callback", yipayHandler.HandleCallback)
	router.GET("/callback.php", yipayHandler.HandleCallback)
	router.POST("/callback.php", yipayHandler.HandleCallback)

	// 签名验证接口 - 支持.php后缀
	router.GET("/api/checksign", yipayHandler.HandleCheckSign)
	router.POST("/api/checksign", yipayHandler.HandleCheckSign)
	router.GET("/api/checksign.php", yipayHandler.HandleCheckSign)
	router.POST("/api/checksign.php", yipayHandler.HandleCheckSign)

	// 系统接口
	router.GET("/health", healthHandler.HandleHealth)
	router.GET("/qrcode", qrcodeHandler.HandleQRCode)
	router.GET("/pay", payHandler.HandlePayPage) // 支付页面（扫码后跳转）

	// WebSocket接口 - 实时订单状态推送
	router.GET("/ws/order", wsHandler.HandleWebSocket)      // 用户支付页面WebSocket
	router.GET("/ws/admin", adminWsHandler.HandleWebSocket) // 管理后台WebSocket

	// 管理后台 - 登录/登出（无需认证）
	router.GET("/admin/login", adminAuth.HandleLogin)
	router.POST("/admin/login", adminAuth.HandleLogin)
	router.GET("/admin/logout", adminAuth.HandleLogout)

	// 管理接口 - 需要认证
	router.GET("/admin/dashboard", adminAuth.RequireAuth(), adminHandler.HandleDashboard) // 管理后台页面
	router.GET("/admin/orders", adminAuth.RequireAuth(), adminHandler.HandleGetOrders)    // 获取订单列表
	router.POST("/admin/action", adminAuth.RequireAuth(), adminHandler.HandleAdminAction) // 新的操作API（基于session）

	// 管理接口 - 兼容旧API（需要pid/key参数）
	router.GET("/admin", adminHandler.HandleAdmin)  // 管理操作API（旧版）
	router.POST("/admin", adminHandler.HandleAdmin) // 管理操作API（旧版）

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// 创建路径规范化的HTTP handler包装器
	// 这个包装器在HTTP层面处理，早于Gin的路由匹配
	pathNormalizingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// 规范化路径：去除多余斜杠
		normalizedPath := path
		for strings.Contains(normalizedPath, "//") {
			normalizedPath = strings.ReplaceAll(normalizedPath, "//", "/")
		}
		// 去除末尾斜杠（保留根路径"/"）
		if len(normalizedPath) > 1 && strings.HasSuffix(normalizedPath, "/") {
			normalizedPath = strings.TrimSuffix(normalizedPath, "/")
		}

		// 更新请求路径
		r.URL.Path = normalizedPath

		// 传递给Gin处理
		router.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      pathNormalizingHandler, // 使用包装后的handler
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	logger.Success("Server starting",
		zap.String("address", addr),
		zap.String("mode", cfg.Server.Mode),
		zap.Bool("http2", true))

	// 优雅退出
	go func() {

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()
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

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// 停止监控服务
	monitorService.Stop()

	logger.Info("Server stopped gracefully")
	logger.Sync()
}
