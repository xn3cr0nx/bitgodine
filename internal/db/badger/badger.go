package badger

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/dgraph-io/badger"
	"github.com/spf13/viper"
)

// Config strcut containing initialization fields
type Config struct {
	Dir string
}

var instance *badger.DB

// Instance creates a new instance of the db
func Instance(conf *Config) (*badger.DB, error) {
	if instance == nil {
		if conf == nil {
			return nil, errors.New("No config provided")
		}

		opts := badger.DefaultOptions
		opts.Dir, opts.ValueDir = conf.Dir, conf.Dir
		db, err := badger.Open(opts)
		if err != nil {
			return nil, err
		}
		instance = db
	}
	return instance, nil
}

// Drop empties the badger store
func Drop() error {
	dir, err := ioutil.ReadDir(viper.GetString("dbDir"))
	if err != nil {
		return err
	}
	for _, d := range dir {
		if err := os.RemoveAll(path.Join([]string{viper.GetString("dbDir"), d.Name()}...)); err != nil {
			return err
		}
	}
	return nil
}
