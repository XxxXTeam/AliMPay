package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

// Init 初始化日志系统
func Init(cfg *Config) error {
	// 设置日志级别
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// 设置编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 选择编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
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

		// 打开日志文件
		logFile, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		fileWriter := zapcore.AddSync(logFile)
		fileCore := zapcore.NewCore(encoder, fileWriter, level)
		cores = append(cores, fileCore)
	}

	// 控制台输出
	if cfg.Output == "stdout" || cfg.Output == "both" {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// 创建logger
	core := zapcore.NewTee(cores...)
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
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
