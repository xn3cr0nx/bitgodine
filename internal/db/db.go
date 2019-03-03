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

// Instance creates a new instance of the db
func Instance(conf *Config) (*database.DB, error) {
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

// GetBlock returnes a *Block looking for the block corresponding to the hash passed
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

// // Store creates a new bucket named with the transaction id and fills it with the corresponding block hash
// // and spent transactions outputs mapped as previous output (txid) and index (vout)
// func (tx *Tx) Store(leveldb *database.DB, blockHash *chainhash.Hash, blockHeight int32) error {
// 	err := (*leveldb).Update(func(t database.Tx) error {
// 		txBucket, err := t.Metadata().CreateBucketIfNotExists([]byte(tx.Hash().String()))
// 		if err != nil {
// 			return err
// 		}

// 		// if err := txBucket.Put([]byte(strconv.Itoa(int(blockHeight))), []byte(blockHash.String())); err != nil {
// 		if err := txBucket.Put([]byte("block"), []byte(blockHash.String())); err != nil {
// 			return err
// 		}

// 		for _, txIn := range tx.MsgTx().TxIn {
// 			if err := txBucket.Put([]byte(txIn.PreviousOutPoint.Hash.String()), []byte(strconv.Itoa(int(txIn.PreviousOutPoint.Index)))); err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // GetTxInputs extracts from the transaction stored the list of map between transaction hash and spent index
// func GetTxInputs(leveldb *database.DB, hash *chainhash.Hash) (map[chainhash.Hash]int32, error) {
// 	var txInputs map[chainhash.Hash]int32

// 	err := (*leveldb).View(func(t database.Tx) error {
// 		txBucket := t.Metadata().Bucket([]byte(hash.String()))
// 		if txBucket == nil {
// 			return errors.New("Tx not found in DB")
// 		}

// 		err := txBucket.ForEach(func(k, v []byte) error {
// 			if string(k) != "block" {
// 				hash, err := chainhash.NewHash(k)
// 				if err != nil {
// 					return err
// 				}
// 				vout, err := strconv.Atoi(string(v))
// 				if err != nil {
// 					return err
// 				}
// 				txInputs[*hash] = int32(vout)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 			return nil
// 		})
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return txInputs, nil
// }

// // GetTxBlock extracts the block hash the transaction in contained in
// func GetTxBlock(leveldb *database.DB, hash *chainhash.Hash) (chainhash.Hash, error) {
// 	var loadedBlockBytes []byte

// 	err := (*leveldb).View(func(t database.Tx) error {
// 		txBucket := t.Metadata().Bucket([]byte(hash.String()))
// 		if txBucket == nil {
// 			return errors.New("Tx not found in DB")
// 		}

// 		cursor := txBucket.Cursor()
// 		if first := cursor.First(); first == false {
// 			return nil
// 		}

// 		blockBytes := cursor.Value()

// 		loadedBlockBytes = make([]byte, len(blockBytes))
// 		copy(loadedBlockBytes, blockBytes)

// 		return nil
// 	})
// 	if err != nil {
// 		return chainhash.Hash{}, err
// 	}

// 	blockHash, err := chainhash.NewHash(loadedBlockBytes)
// 	if err != nil {
// 		return chainhash.Hash{}, err
// 	}

// 	return *blockHash, nil
// }

// func GetSpentTx(txIn *wire.TxIn) (Tx, error) {
// 	leveldb, _ := db.Instance(nil)
// 	blockHash, err := GetTxBlock(leveldb, &txIn.PreviousOutPoint.Hash)
// 	if err != nil {
// 		return Tx{}, err
// 	}

// 	block, err := blocks.Get(leveldb, &blockHash)
// 	if err != nil {
// 		return Tx{}, err
// 	}

// 	transactions := block.Transactions()
// 	for _, t := range transactions {
// 		if t.Hash().IsEqual(&txIn.PreviousOutPoint.Hash) {
// 			return Tx{Tx: *t}, nil
// 		}
// 	}

// 	return Tx{}, nil
// }
