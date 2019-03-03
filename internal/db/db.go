package db

import (
	"errors"
	"path/filepath"

	"github.com/btcsuite/btcd/database"
	_ "github.com/btcsuite/btcd/database/ffldb"
	"github.com/btcsuite/btcd/wire"
)

// Config strcut containing initialization fields
type Config struct {
	Dir  string
	Name string
	Net  wire.BitcoinNet
}

var instance *database.DB

// LevelDB creates a new instance of the db
func LevelDB(conf *Config) (*database.DB, error) {
	if instance == nil {
		if conf == nil {
			return nil, errors.New("No config provided")
		}
		dbPath := filepath.Join(conf.Dir, conf.Name)
		inst, err := database.Create("ffldb", dbPath, conf.Net)
		instance = &inst
		if err != nil {
			if err.Error() == "file already exists: file already exists" {
				inst, err := database.Open("ffldb", dbPath, conf.Net)
				instance = &inst
				if err != nil {
					return nil, err
				}
				return instance, nil
			}
			return nil, err
		}
		return instance, nil
	}

	return instance, nil
}
