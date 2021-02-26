package cache

import (
	"time"

	"github.com/coocood/freecache"
)

type Options struct {
	SizeInMB int `envconfig:"CACHE_SIZE_IN_MB" default:"100"`
}

type Cache struct {
	cacheExecutor *freecache.Cache
}

const (
	MB = 1024 * 1024
)

func NewCache(options Options) *Cache {
	executor := freecache.NewCache(options.SizeInMB * MB)
	return &Cache{cacheExecutor: executor}
}

func (c *Cache) Insert(key []byte, value []byte, ttl time.Duration) error {
	return c.cacheExecutor.Set(key, value, int(ttl/time.Second))
}

func (c *Cache) Fetch(key []byte) ([]byte, error) {
	return c.cacheExecutor.Get(key)
}
