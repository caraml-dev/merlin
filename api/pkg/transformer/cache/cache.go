package cache

import (
	"time"

	"github.com/coocood/freecache"
)

type Cache interface {
	Insert(key []byte, value []byte, ttl time.Duration) error
	Fetch(key []byte) ([]byte, error)
}

type Options struct {
	SizeInMB int `envconfig:"CACHE_SIZE_IN_MB" default:"100"`
}

type inMemoryCache struct {
	cache *freecache.Cache
}

const (
	MB = 1024 * 1024
)

func NewInMemoryCache(options *Options) *inMemoryCache {
	executor := freecache.NewCache(options.SizeInMB * MB)
	return &inMemoryCache{cache: executor}
}

func (c *inMemoryCache) Insert(key []byte, value []byte, ttl time.Duration) error {
	return c.cache.Set(key, value, int(ttl/time.Second))
}

func (c *inMemoryCache) Fetch(key []byte) ([]byte, error) {
	return c.cache.Get(key)
}
