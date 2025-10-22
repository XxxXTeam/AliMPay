package lock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"alimpay-go/pkg/logger"

	"go.uber.org/zap"
)

// FileLock 文件锁
type FileLock struct {
	filePath string
	timeout  time.Duration
	file     *os.File
	mu       sync.Mutex
}

// LockInfo 锁信息
type LockInfo struct {
	Timestamp time.Time `json:"timestamp"`
	Timeout   int64     `json:"timeout"`
	ServerID  string    `json:"server_id"`
}

// NewFileLock 创建文件锁
func NewFileLock(filePath string, timeout time.Duration) *FileLock {
	return &FileLock{
		filePath: filePath,
		timeout:  timeout,
	}
}

// TryLock 尝试获取锁（非阻塞）
func (fl *FileLock) TryLock() (bool, error) {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	// 清理过期的锁
	if err := fl.cleanupExpiredLock(); err != nil {
		logger.Warn("Failed to cleanup expired lock", zap.Error(err))
	}

	// 确保目录存在
	dir := filepath.Dir(fl.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, fmt.Errorf("failed to create lock directory: %w", err)
	}

	// 尝试创建锁文件（O_EXCL 确保原子性）
	file, err := os.OpenFile(fl.filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			// 锁文件已存在，获取锁失败
			return false, nil
		}
		return false, fmt.Errorf("failed to create lock file: %w", err)
	}

	fl.file = file

	// 写入锁信息
	lockInfo := LockInfo{
		Timestamp: time.Now(),
		Timeout:   int64(fl.timeout.Seconds()),
		ServerID:  fmt.Sprintf("%d", os.Getpid()),
	}

	data, err := json.Marshal(lockInfo)
	if err != nil {
		fl.Unlock()
		return false, fmt.Errorf("failed to marshal lock info: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		fl.Unlock()
		return false, fmt.Errorf("failed to write lock info: %w", err)
	}

	logger.Debug("Lock acquired", zap.String("file", fl.filePath))
	return true, nil
}

// Unlock 释放锁
func (fl *FileLock) Unlock() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	if fl.file != nil {
		fl.file.Close()
		fl.file = nil
	}

	// 删除锁文件
	if err := os.Remove(fl.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	logger.Debug("Lock released", zap.String("file", fl.filePath))
	return nil
}

// cleanupExpiredLock 清理过期的锁
func (fl *FileLock) cleanupExpiredLock() error {
	// 检查锁文件是否存在
	info, err := os.Stat(fl.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// 读取锁信息
	data, err := os.ReadFile(fl.filePath)
	if err != nil {
		return err
	}

	var lockInfo LockInfo
	if err := json.Unmarshal(data, &lockInfo); err != nil {
		// 锁信息无效，直接删除
		os.Remove(fl.filePath)
		logger.Warn("Removed invalid lock file", zap.String("file", fl.filePath))
		return nil
	}

	// 检查锁是否过期
	lockAge := time.Since(lockInfo.Timestamp)
	lockTimeout := time.Duration(lockInfo.Timeout) * time.Second

	if lockAge > lockTimeout {
		// 锁已过期，删除
		os.Remove(fl.filePath)
		logger.Info("Removed expired lock file",
			zap.String("file", fl.filePath),
			zap.Duration("age", lockAge),
			zap.Duration("timeout", lockTimeout))
	}

	_ = info // 避免未使用变量警告

	return nil
}

// AmountLock 金额分配锁（用于经营码模式的金额去重）
type AmountLock struct {
	mu sync.Mutex
}

var globalAmountLock = &AmountLock{}

// GetAmountLock 获取全局金额锁
func GetAmountLock() *AmountLock {
	return globalAmountLock
}

// Lock 加锁
func (al *AmountLock) Lock() {
	al.mu.Lock()
}

// Unlock 解锁
func (al *AmountLock) Unlock() {
	al.mu.Unlock()
}
