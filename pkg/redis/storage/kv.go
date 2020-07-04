package storage

import (
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/redis"
)

// KV instance of key value store designed to treat block structs
type KV struct {
	*redis.Redis
	cache *cache.Cache
}

// Config strcut containing initialization fields
type Config struct {
	URL string
}

// Conf returnes default config struct
func Conf(path string) *Config {
	url := viper.GetString("redis")
	if path != "" {
		url = path
	}

	return &Config{
		URL: url,
	}
}

// NewKV creates a new instance of KV
func NewKV(conf *Config, c *cache.Cache) (*KV, error) {
	db, err := redis.NewRedis(&redis.Config{URL: conf.URL})
	if err != nil {
		return nil, err
	}
	return &KV{db, c}, nil
}
