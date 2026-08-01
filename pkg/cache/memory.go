package cache

import (
	"context"
	"sync"
	"time"
)

type memoryItem struct {
	value    string
	expireAt time.Time // 零值表示永不过期
}

func (it memoryItem) expired() bool {
	return !it.expireAt.IsZero() && time.Now().After(it.expireAt)
}

// Memory 进程内内存缓存，带定期清理，适用于单机部署或无 Redis 的开发环境。
type Memory struct {
	mu    sync.RWMutex
	items map[string]memoryItem
	stop  chan struct{}
	once  sync.Once
}

// NewMemory 创建内存缓存并启动每分钟一次的过期清理。
func NewMemory() *Memory {
	m := &Memory{
		items: make(map[string]memoryItem),
		stop:  make(chan struct{}),
	}
	go m.janitor()
	return m
}

func (m *Memory) janitor() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			for k, it := range m.items {
				if it.expired() {
					delete(m.items, k)
				}
			}
			m.mu.Unlock()
		case <-m.stop:
			return
		}
	}
}

// Get 实现 Cache 接口。
func (m *Memory) Get(_ context.Context, key string) (string, error) {
	m.mu.RLock()
	it, ok := m.items[key]
	m.mu.RUnlock()
	if !ok || it.expired() {
		return "", ErrNotFound
	}
	return it.value, nil
}

// Set 实现 Cache 接口。
func (m *Memory) Set(_ context.Context, key, value string, ttl time.Duration) error {
	it := memoryItem{value: value}
	if ttl > 0 {
		it.expireAt = time.Now().Add(ttl)
	}
	m.mu.Lock()
	m.items[key] = it
	m.mu.Unlock()
	return nil
}

// Del 实现 Cache 接口。
func (m *Memory) Del(_ context.Context, keys ...string) error {
	m.mu.Lock()
	for _, k := range keys {
		delete(m.items, k)
	}
	m.mu.Unlock()
	return nil
}

// Exists 实现 Cache 接口。
func (m *Memory) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.Get(ctx, key)
	if err == ErrNotFound {
		return false, nil
	}
	return err == nil, err
}

// Close 停止后台清理协程。
func (m *Memory) Close() error {
	m.once.Do(func() { close(m.stop) })
	return nil
}
