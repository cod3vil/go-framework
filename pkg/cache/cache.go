// Package cache 提供统一缓存接口，内置 Redis 与内存两种实现。
// 未配置 Redis 时框架自动降级为内存缓存，保证零依赖可运行。
package cache

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound 键不存在或已过期。
var ErrNotFound = errors.New("cache: key not found")

// Cache 统一缓存接口。
type Cache interface {
	// Get 获取值；键不存在返回 ErrNotFound。
	Get(ctx context.Context, key string) (string, error)
	// Set 写入值；ttl <= 0 表示永不过期。
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	// Del 删除一个或多个键。
	Del(ctx context.Context, keys ...string) error
	// Exists 判断键是否存在。
	Exists(ctx context.Context, key string) (bool, error)
	// Close 释放底层资源。
	Close() error
}
