package main

import (
	"flag"
	"fmt"
	"os"

	"alimpay-go/internal/config"
	"alimpay-go/internal/database"
	"alimpay-go/pkg/logger"
)

func main() {
	configPath := flag.String("config", "./configs/config.yaml", "Path to configuration file")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志（简化版）
	logCfg := &logger.Config{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}
	logger.Init(logCfg)

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
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("✓ Database initialized successfully!")
	fmt.Printf("  Database file: %s\n", cfg.Database.Path)
}
