package storage

import (
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/tikv"
)

// KV instance of key value store designed to treat block structs
type KV struct {
	*tikv.TiKV
	cache *cache.Cache
}

// Config strcut containing initialization fields
type Config struct {
	Dir string
}

// Conf returnes default config struct
func Conf(path string) *Config {
	dir := viper.GetString("db")
	if path != "" {
		dir = path
	}

	return &Config{
		Dir: dir,
	}
}

// NewKV creates a new instance of KV
func NewKV(conf *Config, c *cache.Cache) (*KV, error) {
	db, err := tikv.NewTiKV(&tikv.Config{})
	if err != nil {
		return nil, err
	}
	return &KV{db, c}, nil
}
