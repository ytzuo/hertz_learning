package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrCacheMiss = errors.New("cache miss")

type entry struct {
	value     any
	expiresAt time.Time
}

type MemoryRedis struct {
	mu    sync.RWMutex
	items map[string]entry
}

// MemoryRedis 是本地 demo 使用的 Redis-like 缓存适配器。
// 后续可以替换成真实 Redis 客户端，而无需改动 service 层。
func NewMemoryRedis() *MemoryRedis {
	return &MemoryRedis{
		items: map[string]entry{},
	}
}

func (r *MemoryRedis) Get(ctx context.Context, key string) (any, error) {
	r.mu.RLock()
	item, ok := r.items[key]
	r.mu.RUnlock()

	if !ok || expired(item) {
		return nil, ErrCacheMiss
	}
	return item.value, nil
}

func (r *MemoryRedis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[key] = entry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (r *MemoryRedis) Delete(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.items, key)
	return nil
}

func expired(item entry) bool {
	return !item.expiresAt.IsZero() && time.Now().After(item.expiresAt)
}
