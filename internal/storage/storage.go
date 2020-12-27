package storage

import (
	"github.com/spf13/viper"

	"github.com/xn3cr0nx/bitgodine/internal/storage/badger"
	"github.com/xn3cr0nx/bitgodine/internal/storage/redis"
	"github.com/xn3cr0nx/bitgodine/internal/storage/tikv"
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

// NewStorage returns a new DB instance based on the environment
func NewStorage() (db DB, err error) {
	switch viper.GetString("db") {
	case "tikv":
		db, err = tikv.NewTiKV(tikv.Conf(viper.GetString("tikv")))
		if err != nil {
			return
		}
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
