package cache

import (
	"time"

	cache "github.com/patrickmn/go-cache"
)

type Cache interface {
	Insert(key string, value interface{}, ttl time.Duration)
	Fetch(key string) (interface{}, bool)
}

type inMemoryCache struct {
	cache *cache.Cache
}

const (
	cacheCleanUpSeconds = 300
)

func NewInMemoryCache() *inMemoryCache {
	executor := cache.New(0, cacheCleanUpSeconds*time.Second)
	return &inMemoryCache{cache: executor}
}

func (c *inMemoryCache) Insert(key string, value interface{}, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

func (c *inMemoryCache) Fetch(key string) (interface{}, bool) {
	return c.cache.Get(key)
}
