package kv

import (
	"github.com/spf13/viper"

	"github.com/xn3cr0nx/bitgodine/internal/storage/kv/badger"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv/redis"
)

// DB interface implements methods for a generic key value storage db
type DB interface {
	Store(string, []byte) error
	StoreBatch(interface{}) error
	StoreQueueBatch(interface{}) error
	Read(string) ([]byte, error)
	ReadKeys() ([]string, error)
	ReadKeyValues() (map[string][]byte, error)
	ReadKeysWithPrefix(string) ([]string, error)
	ReadPrefix(string) ([][]byte, error)
	ReadFirstValueByPrefix(string) ([]byte, error)
	ReadPrefixWithKey(string) (map[string][]byte, error)
	IsStored(string) bool
	Delete(string) error
	Empty() error
	Close() error
}

// NewDB returns a new DB instance based on the environment
func NewDB() (db DB, err error) {
	switch viper.GetString("db") {
	// case "tikv":
	// 	db, err = tikv.NewTiKV(tikv.Conf(viper.GetString("tikv")))
	// 	if err != nil {
	// 		return
	// 	}
	case "badger":
		db, err = badger.NewBadger(badger.Conf(viper.GetString("badger")), false)
		if err != nil {
			return
		}
	case "redis":
		db, err = redis.NewRedis(redis.Conf(viper.GetString("redis")))
		if err != nil {
			return
		}
	}
	return
}
