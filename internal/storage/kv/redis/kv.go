package redis

import (
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// KV instance of key value store designed to treat block structs
type KV struct {
	*Redis
	cache *cache.Cache
}

// NewKV creates a new instance of KV
func NewKV(r *Redis, c *cache.Cache) (*KV, error) {
	return &KV{r, c}, nil
}
