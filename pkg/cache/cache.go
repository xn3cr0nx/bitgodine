package cache

import (
	"github.com/dgraph-io/ristretto"
)

// Cache wrapper on ristretto cache
type Cache struct {
	*ristretto.Cache
}

// NewCache creates a new instance of the cache
func NewCache(conf *ristretto.Cache) (*Cache, error) {
	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000000 * 10,
		MaxCost:     1000000,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	return &Cache{c}, nil
}
