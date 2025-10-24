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
	Enabled        bool     `yaml:"enabled"`
	QRCodePath     string   `yaml:"qr_code_path"`  // 单个二维码路径（向后兼容）
	QRCodePaths    []QRCode `yaml:"qr_code_paths"` // 多个二维码配置
	QRCodeID       string   `yaml:"qr_code_id"`    // 支付宝收款码ID，用于手机端拉起支付宝（单个模式）
	AmountOffset   float64  `yaml:"amount_offset"`
	MatchTolerance int      `yaml:"match_tolerance"`
	PaymentTimeout int      `yaml:"payment_timeout"`
	PollingMode    string   `yaml:"polling_mode"` // 轮询模式: round_robin, random, least_used
}

// QRCode 二维码配置
type QRCode struct {
	ID       string `yaml:"id"`       // 二维码唯一标识
	Path     string `yaml:"path"`     // 二维码图片路径
	CodeID   string `yaml:"code_id"`  // 支付宝收款码ID
	Enabled  bool   `yaml:"enabled"`  // 是否启用
	Priority int    `yaml:"priority"` // 优先级（数字越小优先级越高）

	// 独立的支付宝API配置（可选，为空则使用全局配置）
	AlipayAPI *QRCodeAlipayConfig `yaml:"alipay_api,omitempty"`
}

// QRCodeAlipayConfig 二维码专属的支付宝API配置
type QRCodeAlipayConfig struct {
	ServerURL       string `yaml:"server_url,omitempty"`        // 支付宝网关
	AppID           string `yaml:"app_id,omitempty"`            // 应用ID
	PrivateKey      string `yaml:"private_key,omitempty"`       // 应用私钥
	AlipayPublicKey string `yaml:"alipay_public_key,omitempty"` // 支付宝公钥
	TransferUserID  string `yaml:"transfer_user_id,omitempty"`  // 转账用户ID
	SignType        string `yaml:"sign_type,omitempty"`         // 签名类型
	Charset         string `yaml:"charset,omitempty"`           // 字符集
	Format          string `yaml:"format,omitempty"`            // 格式
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

	// 设置默认轮询模式
	if cfg.Payment.BusinessQRMode.PollingMode == "" {
		cfg.Payment.BusinessQRMode.PollingMode = "round_robin"
	}

	// 如果配置了单个二维码路径但没有配置多个二维码，自动转换为多二维码模式
	if cfg.Payment.BusinessQRMode.QRCodePath != "" && len(cfg.Payment.BusinessQRMode.QRCodePaths) == 0 {
		cfg.Payment.BusinessQRMode.QRCodePaths = []QRCode{
			{
				ID:       "default",
				Path:     cfg.Payment.BusinessQRMode.QRCodePath,
				CodeID:   cfg.Payment.BusinessQRMode.QRCodeID,
				Enabled:  true,
				Priority: 1,
			},
		}
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

// GetEffectiveAlipayConfig 获取二维码的有效支付宝配置（如果二维码有独立配置则使用，否则使用全局配置）
func (qr *QRCode) GetEffectiveAlipayConfig(globalConfig *AlipayConfig) *AlipayConfig {
	// 如果没有独立配置，直接返回全局配置
	if qr.AlipayAPI == nil {
		return globalConfig
	}

	// 创建合并后的配置（二维码配置优先，缺失部分使用全局配置）
	merged := &AlipayConfig{
		ServerURL:       qr.AlipayAPI.ServerURL,
		AppID:           qr.AlipayAPI.AppID,
		PrivateKey:      qr.AlipayAPI.PrivateKey,
		AlipayPublicKey: qr.AlipayAPI.AlipayPublicKey,
		TransferUserID:  qr.AlipayAPI.TransferUserID,
		SignType:        qr.AlipayAPI.SignType,
		Charset:         qr.AlipayAPI.Charset,
		Format:          qr.AlipayAPI.Format,
	}

	// 填充缺失的字段
	if merged.ServerURL == "" {
		merged.ServerURL = globalConfig.ServerURL
	}
	if merged.AppID == "" {
		merged.AppID = globalConfig.AppID
	}
	if merged.PrivateKey == "" {
		merged.PrivateKey = globalConfig.PrivateKey
	}
	if merged.AlipayPublicKey == "" {
		merged.AlipayPublicKey = globalConfig.AlipayPublicKey
	}
	if merged.TransferUserID == "" {
		merged.TransferUserID = globalConfig.TransferUserID
	}
	if merged.SignType == "" {
		merged.SignType = globalConfig.SignType
	}
	if merged.Charset == "" {
		merged.Charset = globalConfig.Charset
	}
	if merged.Format == "" {
		merged.Format = globalConfig.Format
	}

	return merged
}

// HasIndependentAPI 检查二维码是否配置了独立的API
func (qr *QRCode) HasIndependentAPI() bool {
	return qr.AlipayAPI != nil && qr.AlipayAPI.AppID != ""
}
