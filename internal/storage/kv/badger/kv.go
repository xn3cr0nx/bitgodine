package badger

import (
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// KV instance of key value store designed to treat block structs
type KV struct {
	*Badger
	cache *cache.Cache
}

// NewKV creates a new instance of KV
func NewKV(bdg *Badger, c *cache.Cache) (*KV, error) {
	return &KV{bdg, c}, nil
}
