package cache

import (
	"errors"

	"github.com/allegro/bigcache"
)

var instance *bigcache.BigCache

// Instance creates a new instance of the cache
func Instance(conf bigcache.Config) (*bigcache.BigCache, error) {
	if instance == nil {
		if conf.LifeWindow == 0 {
			return nil, errors.New("No config provided")
		}

		cache, err := bigcache.NewBigCache(conf)
		if err != nil {
			return nil, err
		}
		instance = cache
	}
	return instance, nil
}

// Drop empties the badger store
func Drop() error {
	if err := instance.Reset(); err != nil {
		return err
	}
	return nil
}
