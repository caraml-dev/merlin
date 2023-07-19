package cache

import (
	"time"

	"github.com/coocood/freecache"
)

type Cache interface {
	Insert(key []byte, value []byte, ttl time.Duration) error
	Fetch(key []byte) ([]byte, error)
}

type inMemoryCache struct {
	cache *freecache.Cache
}

const (
	MB = 1024 * 1024
)

func NewInMemoryCache(sizeInMB int) *inMemoryCache {
	executor := freecache.NewCache(sizeInMB * MB)
	return &inMemoryCache{cache: executor}
}

func (c *inMemoryCache) Insert(key []byte, value []byte, ttl time.Duration) error {
	return c.cache.Set(key, value, int(ttl/time.Second))
}

func (c *inMemoryCache) Fetch(key []byte) ([]byte, error) {
	return c.cache.Get(key)
}
