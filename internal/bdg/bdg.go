package bdg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/badger"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
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

// StoreBlock inserts in the db the block as []byte passed
func StoreBlock(b *blocks.Block) error {
	if IsStored(b.Hash()) {
		return errors.New(fmt.Sprintf("block %s already exists", b.Hash().String()))
	}
	return instance.Update(func(txn *badger.Txn) error {
		buff := new(bytes.Buffer)
		serial := bufio.NewWriter(buff)
		b.MsgBlock().Serialize(serial)
		serial.Flush()
		return txn.Set(b.Hash().CloneBytes(), buff.Bytes())
	})
}

// GetBlock returnes a *Block looking for the block corresponding to the hash passed
func GetBlock(hash *chainhash.Hash) (*blocks.Block, error) {
	var loadedBlockBytes []byte
	err := instance.View(func(txn *badger.Txn) error {
		item, err := txn.Get(hash.CloneBytes())
		if err != nil {
			fmt.Println("error", err)
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		loadedBlockBytes = make([]byte, len(val))
		copy(loadedBlockBytes, val)
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

// IsStored returns true if the block corresponding to passed hash is stored in db
func IsStored(hash *chainhash.Hash) bool {
	err := instance.View(func(txn *badger.Txn) error {
		_, err := txn.Get(hash.CloneBytes())
		if err != nil {
			return err
		}
		return nil
	})
	return err == nil
}

// StoredBlocks is an utility functions that returnes the list of stored blocks hash
func StoredBlocks() ([]string, error) {
	var blocks []string
	err := instance.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			hash, err := chainhash.NewHash(k)
			if err != nil {
				return err
			}
			blocks = append(blocks, hash.String())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

// // Drop empties the badger store
// func Drop() error {
// 	return instance.DropAll()
// }
