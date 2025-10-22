package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Alipay   AlipayConfig   `yaml:"alipay"`
	Database DatabaseConfig `yaml:"database"`
	Payment  PaymentConfig  `yaml:"payment"`
	Merchant MerchantConfig `yaml:"merchant"`
	Logging  LoggingConfig  `yaml:"logging"`
	Monitor  MonitorConfig  `yaml:"monitor"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Mode         string `yaml:"mode"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	BaseURL      string `yaml:"base_url"` // 基础URL，留空则自动获取
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	ServerURL       string `yaml:"server_url"`
	AppID           string `yaml:"app_id"`
	PrivateKey      string `yaml:"private_key"`
	AlipayPublicKey string `yaml:"alipay_public_key"`
	TransferUserID  string `yaml:"transfer_user_id"`
	SignType        string `yaml:"sign_type"`
	Charset         string `yaml:"charset"`
	Format          string `yaml:"format"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            string `yaml:"type"`
	Path            string `yaml:"path"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// PaymentConfig 支付配置
type PaymentConfig struct {
	MaxWaitTime      int               `yaml:"max_wait_time"`
	CheckInterval    int               `yaml:"check_interval"`
	QueryMinutesBack int               `yaml:"query_minutes_back"`
	OrderTimeout     int               `yaml:"order_timeout"`
	AutoCleanup      bool              `yaml:"auto_cleanup"`
	QRCodeSize       int               `yaml:"qr_code_size"`
	QRCodeMargin     int               `yaml:"qr_code_margin"`
	BusinessQRMode   BusinessQRMode    `yaml:"business_qr_mode"`
	AntiRiskURL      AntiRiskURLConfig `yaml:"anti_risk_url"`
}

// BusinessQRMode 经营码收款模式配置
type BusinessQRMode struct {
	Enabled        bool    `yaml:"enabled"`
	QRCodePath     string  `yaml:"qr_code_path"`
	AmountOffset   float64 `yaml:"amount_offset"`
	MatchTolerance int     `yaml:"match_tolerance"`
	PaymentTimeout int     `yaml:"payment_timeout"`
}

// AntiRiskURLConfig 防风控URL配置
type AntiRiskURLConfig struct {
	Enabled           bool   `yaml:"enabled"`
	OuterAppID        string `yaml:"outer_app_id"`
	InnerAppID        string `yaml:"inner_app_id"`
	MdeductLandingURL string `yaml:"mdeduct_landing_url"`
	RenderSchemeURL   string `yaml:"render_scheme_url"`
}

// MerchantConfig 商户配置
type MerchantConfig struct {
	ID   string `yaml:"id"`
	Key  string `yaml:"key"`
	Rate int    `yaml:"rate"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	Enabled     bool `yaml:"enabled"`
	Interval    int  `yaml:"interval"`
	LockTimeout int  `yaml:"lock_timeout"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认值
	setDefaults(&cfg)

	// 验证配置
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// setDefaults 设置默认值
func setDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 60
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 60
	}

	if cfg.Database.Type == "" {
		cfg.Database.Type = "sqlite3"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./data/alimpay.db"
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 100
	}

	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "file"
	}
	if cfg.Logging.FilePath == "" {
		cfg.Logging.FilePath = "./logs/alimpay.log"
	}

	if cfg.Payment.QRCodeSize == 0 {
		cfg.Payment.QRCodeSize = 300
	}
	if cfg.Payment.QRCodeMargin == 0 {
		cfg.Payment.QRCodeMargin = 10
	}
}

// validate 验证配置
func validate(cfg *Config) error {
	// 创建必要的目录
	dirs := []string{
		filepath.Dir(cfg.Database.Path),
		filepath.Dir(cfg.Logging.FilePath),
		"./qrcode",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Save 保存配置到文件
func Save(cfg *Config, configPath string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
