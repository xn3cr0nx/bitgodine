package test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/pkg/badger/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/storage"
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

func InitDB() (db storage.DB, err error) {
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
	db, err = kv.NewKV(kv.Conf(filepath.Join(bitgodineFolder, "badger")), ca, false)
	if err != nil {
		return
	}
	return
}

func CleanTestDB(db storage.DB) (err error) {
	DB := db.(*kv.KV)
	err = (*DB).Close()
	if err != nil {
		return
	}
	err = os.RemoveAll(filepath.Join(".", "test"))
	return
}
