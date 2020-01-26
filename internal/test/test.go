package test

import (
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger/kv"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/cache"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
)

func InitTestDB() (db storage.DB, err error) {
	conf := &kv.Config{
		Dir: filepath.Join(".", "test"),
	}
	ca, err := cache.NewCache(nil)
	if err != nil {
		return
	}
	db, err = kv.NewKV(conf, ca, false)
	if err != nil {
		return
	}
	return
}

func CleanTestDB(db storage.DB) (err error) {
	DB := db.(*kv.KV)
	viper.SetDefault("dbDir", filepath.Join(".", "test"))
	err = (*DB).Close()
	if err != nil {
		return
	}
	err = (*DB).Empty()
	return
}
