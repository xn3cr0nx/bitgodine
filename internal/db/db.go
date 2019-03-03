package db

import (
	"errors"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/database"
	_ "github.com/btcsuite/btcd/database/ffldb"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
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

// Get returnes a *Block looking for the block corresponding to the hash passed
func GetBlock(hash *chainhash.Hash) (*blocks.Block, error) {
	var loadedBlockBytes []byte
	err := (*instance).View(func(tx database.Tx) error {
		blockBytes, err := tx.FetchBlock(hash)
		if err != nil {
			return err
		}

		loadedBlockBytes = make([]byte, len(blockBytes))
		copy(loadedBlockBytes, blockBytes)
		return nil
	})
	if err != nil {
		return nil, err
	}

	block, err := btcutil.NewBlockFromBytes(loadedBlockBytes)
	if err != nil {
		return nil, err
	}

	return &blocks.Block{Block: *block}, nil
}

// StoreBlock inserts in the db the block as []byte passed
func StoreBlock(b *blocks.Block) error {
	err := (*instance).Update(func(tx database.Tx) error {
		return tx.StoreBlock(btcutil.NewBlock(b.MsgBlock()))
	})
	if err != nil {
		return err
	}

	return nil
}
