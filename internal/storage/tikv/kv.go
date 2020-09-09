package tikv

import (
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// KV instance of key value store designed to treat block structs
type KV struct {
	*TiKV
	cache *cache.Cache
}

// NewKV creates a new instance of KV
func NewKV(t *TiKV, c *cache.Cache) (*KV, error) {
	return &KV{t, c}, nil
}
