package test

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv/badger"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// InitTestDB setup badger db for test
func InitTestDB() (db kv.DB, err error) {
	conf := &badger.Config{
		Dir: filepath.Join(".", "test"),
	}
	ca, err := cache.NewCache(nil)
	if err != nil {
		return
	}

	bdg, err := badger.NewBadger(conf, false)
	db, err = badger.NewKV(bdg, ca)
	if err != nil {
		return
	}
	return
}

// InitDB setup badger db for test
func InitDB() (db kv.DB, err error) {
	viper.SetDefault("dbDir", filepath.Join(".", "test"))
	hd, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("Bitgodine %v", err))
	}
	bitgodineFolder := filepath.Join(hd, ".bitgodine")
	ca, err := cache.NewCache(nil)
	if err != nil {
		return
	}
	bdg, err := badger.NewBadger(badger.Conf(filepath.Join(bitgodineFolder, "badger")), false)
	db, err = badger.NewKV(bdg, ca)
	if err != nil {
		return
	}
	return
}

// CleanTestDB cleanup badger db for test
func CleanTestDB(db kv.DB) (err error) {
	return db.Empty()
}
