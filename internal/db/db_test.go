package db

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/database"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

func TestDB(t *testing.T) {
	conf := &DBConfig{
		"/tmp", "test", wire.MainNet,
	}
	db, err := DB(conf)
	if err != nil {
		logger.Error("sync", err, logger.Params{})
		return
	}
	defer os.RemoveAll(filepath.Join(conf.Dir, conf.Name))
	defer (*db).Close()

	// Use the Update function of the database to perform a managed
	// read-write transaction and store a genesis block in the database as
	// and example.
	err = (*db).Update(func(tx database.Tx) error {
		genesisBlock := chaincfg.MainNetParams.GenesisBlock
		return tx.StoreBlock(btcutil.NewBlock(genesisBlock))
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Use the View function of the database to perform a managed read-only
	// transaction and fetch the block stored above.
	var loadedBlockBytes []byte
	err = (*db).Update(func(tx database.Tx) error {
		genesisHash := chaincfg.MainNetParams.GenesisHash
		blockBytes, err := tx.FetchBlock(genesisHash)
		if err != nil {
			return err
		}

		// As documented, all data fetched from the database is only
		// valid during a database transaction in order to support
		// zero-copy backends.  Thus, make a copy of the data so it
		// can be used outside of the transaction.
		loadedBlockBytes = make([]byte, len(blockBytes))
		copy(loadedBlockBytes, blockBytes)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := btcutil.NewBlockFromBytes(loadedBlockBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("block", res.Hash())

}
