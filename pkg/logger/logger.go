// Package logger 日志管理包
// @author AliMPay Team
// @description 提供统一的日志管理功能，支持日志轮换、彩色输出等
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	globalLogger *zap.Logger
	sugarLogger  *zap.SugaredLogger
)

// Config 日志配置
type Config struct {
	Level      string
	Format     string
	Output     string
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// 颜色定义
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
	colorWhite  = "\033[97m"

	// 粗体颜色
	colorBoldRed    = "\033[1;31m"
	colorBoldGreen  = "\033[1;32m"
	colorBoldYellow = "\033[1;33m"
	colorBoldBlue   = "\033[1;34m"
	colorBoldPurple = "\033[1;35m"
	colorBoldCyan   = "\033[1;36m"
)

// customColorLevelEncoder 自定义彩色日志级别编码器
func customColorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch level {
	case zapcore.DebugLevel:
		enc.AppendString(colorBoldBlue + "DEBUG" + colorReset)
	case zapcore.InfoLevel:
		enc.AppendString(colorBoldGreen + "INFO " + colorReset)
	case zapcore.WarnLevel:
		enc.AppendString(colorBoldYellow + "WARN " + colorReset)
	case zapcore.ErrorLevel:
		enc.AppendString(colorBoldRed + "ERROR" + colorReset)
	case zapcore.FatalLevel:
		enc.AppendString(colorBoldPurple + "FATAL" + colorReset)
	default:
		enc.AppendString(level.CapitalString())
	}
}

// customTimeEncoder 自定义时间编码器
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(colorCyan + t.Format("2006-01-02 15:04:05.000") + colorReset)
}

// customCallerEncoder 自定义调用者编码器
func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(colorGray + caller.TrimmedPath() + colorReset)
}

// Init 初始化日志系统
func Init(cfg *Config) error {
	// 设置日志级别
	level := zapcore.InfoLevel
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn", "warning":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	}

	// 文件输出的编码器配置（JSON格式）
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 控制台输出的编码器配置（彩色格式）
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:          "T",
		LevelKey:         "L",
		NameKey:          "N",
		CallerKey:        "C",
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "M",
		StacktraceKey:    "S",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      customColorLevelEncoder,
		EncodeTime:       customTimeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     customCallerEncoder,
		ConsoleSeparator: " ",
	}

	// 设置输出
	var cores []zapcore.Core

	// 文件输出
	if cfg.Output == "file" || cfg.Output == "both" {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// 使用lumberjack实现日志轮换
		// 当日志文件达到MaxSize时自动切割
		// 保留最多MaxBackups个备份文件
		// 文件保留MaxAge天后自动删除
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,    // MB
			MaxBackups: cfg.MaxBackups, // 保留备份数
			MaxAge:     cfg.MaxAge,     // 天数
			Compress:   cfg.Compress,   // 是否压缩
			LocalTime:  true,           // 使用本地时间
		}

		// 文件使用JSON格式，便于解析
		fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
		fileWriter := zapcore.AddSync(lumberJackLogger)
		fileCore := zapcore.NewCore(fileEncoder, fileWriter, level)
		cores = append(cores, fileCore)
	}

	// 控制台输出
	if cfg.Output == "stdout" || cfg.Output == "both" {
		// 控制台使用彩色格式，便于查看
		consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// 创建logger
	core := zapcore.NewTee(cores...)
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	sugarLogger = globalLogger.Sugar()

	return nil
}

// GetLogger 获取全局logger
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		// 如果没有初始化，使用默认配置
		logger, _ := zap.NewProduction()
		globalLogger = logger
	}
	return globalLogger
}

// GetSugar 获取sugar logger
func GetSugar() *zap.SugaredLogger {
	if sugarLogger == nil {
		sugarLogger = GetLogger().Sugar()
	}
	return sugarLogger
}

// Info 记录info级别日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Debug 记录debug级别日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Warn 记录warn级别日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 记录error级别日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 记录fatal级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Infof 格式化info日志
func Infof(template string, args ...interface{}) {
	GetSugar().Infof(template, args...)
}

// Debugf 格式化debug日志
func Debugf(template string, args ...interface{}) {
	GetSugar().Debugf(template, args...)
}

// Warnf 格式化warn日志
func Warnf(template string, args ...interface{}) {
	GetSugar().Warnf(template, args...)
}

// Errorf 格式化error日志
func Errorf(template string, args ...interface{}) {
	GetSugar().Errorf(template, args...)
}

// Fatalf 格式化fatal日志并退出程序
func Fatalf(template string, args ...interface{}) {
	GetSugar().Fatalf(template, args...)
}

// Sync 同步日志
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// WithColor 为文本添加颜色
func WithColor(color, text string) string {
	return color + text + colorReset
}

// Success 记录成功信息（绿色）
func Success(msg string, fields ...zap.Field) {
	GetLogger().Info(WithColor(colorGreen, "✓ "+msg), fields...)
}

// Progress 记录进度信息（蓝色）
func Progress(msg string, fields ...zap.Field) {
	GetLogger().Info(WithColor(colorBlue, "➜ "+msg), fields...)
}

// Highlight 记录高亮信息（黄色）
func Highlight(msg string, fields ...zap.Field) {
	GetLogger().Info(WithColor(colorYellow, "★ "+msg), fields...)
}

// JSON 格式化输出JSON对象（用于调试）
func JSON(msg string, data interface{}) {
	GetLogger().Debug(msg, zap.Any("data", data))
}

// Request 记录HTTP请求日志（精简版）
func Request(method, path, ip string, statusCode int, duration float64) {
	statusColor := colorGreen
	if statusCode >= 400 && statusCode < 500 {
		statusColor = colorYellow
	} else if statusCode >= 500 {
		statusColor = colorRed
	}

	GetLogger().Info("",
		zap.String("method", WithColor(colorBlue, method)),
		zap.String("path", path),
		zap.String("ip", colorGray+ip+colorReset),
		zap.String("status", WithColor(statusColor, fmt.Sprintf("%d", statusCode))),
		zap.Float64("ms", duration),
	)
}
