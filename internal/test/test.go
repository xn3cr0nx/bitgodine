package test

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	badgerStorage "github.com/xn3cr0nx/bitgodine/pkg/badger/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/storage"
)

// InitTestDB setup badger db for test
func InitTestDB() (db storage.DB, err error) {
	conf := &badgerStorage.Config{
		Dir: filepath.Join(".", "test"),
	}
	ca, err := cache.NewCache(nil)
	if err != nil {
		return
	}
	db, err = badgerStorage.NewKV(conf, ca, false)
	if err != nil {
		return
	}
	return
}

// InitDB setup badger db for test
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
	db, err = badgerStorage.NewKV(badgerStorage.Conf(filepath.Join(bitgodineFolder, "badger")), ca, false)
	if err != nil {
		return
	}
	return
}

// CleanTestDB cleanup badger db for test
func CleanTestDB(db storage.DB) (err error) {
	return db.Empty()
}
