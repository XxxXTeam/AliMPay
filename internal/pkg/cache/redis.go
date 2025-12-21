package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis缓存管理器
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	ctx := context.Background()

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		// Redis不可用时返回nil（降级为无缓存模式）
		return nil, err
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// Get 获取缓存
func (r *RedisCache) Get(key string) (string, error) {
	if r == nil || r.client == nil {
		return "", redis.Nil
	}
	return r.client.Get(r.ctx, key).Result()
}

// Set 设置缓存
func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	if r == nil || r.client == nil {
		return nil // 降级：不缓存
	}
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Del 删除缓存
func (r *RedisCache) Del(keys ...string) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Del(r.ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(keys ...string) (int64, error) {
	if r == nil || r.client == nil {
		return 0, nil
	}
	return r.client.Exists(r.ctx, keys...).Result()
}

// Close 关闭连接
func (r *RedisCache) Close() error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Close()
}

// IsAvailable 检查Redis是否可用
func (r *RedisCache) IsAvailable() bool {
	return r != nil && r.client != nil
}
