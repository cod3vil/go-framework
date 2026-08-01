package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/redis/go-redis/v9"
)

// Redis 基于 go-redis v9 的缓存实现。
type Redis struct {
	client *redis.Client
}

// NewRedis 连接 Redis 并做连通性检测。
func NewRedis(cfg config.RedisConfig) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("连接 Redis %s 失败: %w", cfg.Addr, err)
	}
	return &Redis{client: client}, nil
}

// Client 暴露底层客户端，供需要高级命令（如分布式锁）的场景使用。
func (r *Redis) Client() *redis.Client { return r.client }

// Get 实现 Cache 接口。
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return val, err
}

// Set 实现 Cache 接口。
func (r *Redis) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 0 // go-redis 中 0 表示不过期
	}
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Del 实现 Cache 接口。
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

// Exists 实现 Cache 接口。
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Close 关闭连接。
func (r *Redis) Close() error { return r.client.Close() }
