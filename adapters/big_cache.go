package adapters

import (
	"context"
	"log"

	"github.com/allegro/bigcache/v3"
	"github.com/kevin07696/receipt-processor/domain"
)

type BigCache struct {
	cache *bigcache.BigCache
}

func NewBigCache(cache *bigcache.BigCache) BigCache {
	return BigCache{cache: cache}
}

func (c *BigCache) Set(ctx context.Context, key string, value interface{}) domain.StatusCode {
	valueString := value.(string)
	valBytes := []byte(valueString)
	if err := c.cache.Set(key, valBytes); err != nil {
		log.Fatalf("Failed to store value: %v: %v", value, err)
	}
	return domain.StatusOK
}

func (c BigCache) Get(ctx context.Context, key string) (interface{}, domain.StatusCode) {
	value, err := c.cache.Get(key)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return value, domain.StatusOK
}
