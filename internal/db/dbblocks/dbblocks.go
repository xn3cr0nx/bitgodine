package dbblocks

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/badger"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// DbBlocks instance of key value store designed to treat block structs
type DbBlocks struct {
	*badger.DB
}

// NewDbBlocks creates a new instance of DbBlocks
func NewDbBlocks(conf *db.Config) (*DbBlocks, error) {
	opts := badger.DefaultOptions
	opts.Dir, opts.ValueDir = conf.Dir, conf.Dir
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &DbBlocks{db}, nil
}

// StoreBlock inserts in the db the block as []byte passed
func (db *DbBlocks) StoreBlock(b *blocks.Block) error {
	// block validation
	if db.IsStored(b.Hash()) {
		return errors.New(fmt.Sprintf("block %s already exists", b.Hash().String()))
	}
	err := db.Update(func(txn *badger.Txn) error {
		bytes, err := b.Bytes()
		if err != nil {
			return err
		}
		return txn.Set(b.Hash().CloneBytes(), bytes)
	})
	return err
}

// StoreBlockPrevHash inserts in the db the block as []byte passed, using the previous hash as key
func (db *DbBlocks) StoreBlockPrevHash(b *blocks.Block) error {
	return db.Update(func(txn *badger.Txn) error {
		bytes, err := b.Bytes()
		if err != nil {
			return err
		}
		return txn.Set(b.MsgBlock().Header.PrevBlock.CloneBytes(), bytes)
	})
}

// GetBlock returnes a *Block looking for the block corresponding to the hash passed
func (db *DbBlocks) GetBlock(hash *chainhash.Hash) (*blocks.Block, error) {
	var loadedBlockBytes []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(hash.CloneBytes())
		if err != nil {
			logger.Debug("DB", fmt.Sprintf("GetBlock %s", err.Error()), logger.Params{})
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
func (db *DbBlocks) IsStored(hash *chainhash.Hash) bool {
	err := db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(hash.CloneBytes())
		if err != nil {
			return err
		}
		return nil
	})
	return err == nil
}

// StoredBlocks is an utility functions that returnes the list of stored blocks hash
func (db *DbBlocks) StoredBlocks() ([]string, error) {
	var blocks []string
	err := db.View(func(txn *badger.Txn) error {
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

// GetAll returnes all the blocks stored in badger
func (db *DbBlocks) GetAll() ([]blocks.Block, error) {
	var storedBlocks []blocks.Block
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			block, err := item.Value()
			if err != nil {
				return err
			}
			b, err := btcutil.NewBlockFromBytes(block)
			if err != nil {
				return err
			}
			storedBlocks = append(storedBlocks, blocks.Block{Block: *b})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return storedBlocks, nil
}

// DeleteBlock inserts in the db the block as []byte passed
func (db *DbBlocks) DeleteBlock(hash *chainhash.Hash) error {
	err := db.Update(func(txn *badger.Txn) error {
		return txn.Delete(hash.CloneBytes())
	})
	return err
}
