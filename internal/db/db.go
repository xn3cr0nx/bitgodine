package db

import (
	"errors"
	"path/filepath"

	"github.com/btcsuite/btcd/database"
	_ "github.com/btcsuite/btcd/database/ffldb"
	"github.com/btcsuite/btcd/wire"
)

type DBConfig struct {
	Dir  string
	Name string
	Net  wire.BitcoinNet
}

var db *database.DB

func DB(conf *DBConfig) (*database.DB, error) {
	if conf == nil {
		return nil, errors.New("No config provided")
	}

	if db != nil {
		return db, nil
	}

	dbPath := filepath.Join(conf.Dir, conf.Name)
	db, err := database.Create("ffldb", dbPath, conf.Net)
	if err != nil {
		if err.Error() == "file already exists: file already exists" {
			db, err := database.Open("ffldb", dbPath, conf.Net)
			if err != nil {
				return nil, err
			}
			return &db, nil
		}
		return nil, err
	}

	return &db, nil
}
