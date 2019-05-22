package db

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/badger"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
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
	err := instance.Update(func(txn *badger.Txn) error {
		buff := new(bytes.Buffer)
		serial := bufio.NewWriter(buff)
		b.MsgBlock().Serialize(serial)

		bnr := make([]byte, 4)
		binary.LittleEndian.PutUint32(bnr, uint32(b.Height()))
		buff.Write(bnr)
		serial.Flush()
		return txn.Set(b.Hash().CloneBytes(), buff.Bytes())
	})
	return err
}

// StoreBlockPrevHash inserts in the db the block as []byte passed, using the previous hash as key
func StoreBlockPrevHash(b *blocks.Block) error {
	return instance.Update(func(txn *badger.Txn) error {
		buff := new(bytes.Buffer)
		serial := bufio.NewWriter(buff)
		b.MsgBlock().Serialize(serial)

		bnr := make([]byte, 4)
		binary.LittleEndian.PutUint32(bnr, uint32(b.Height()))
		buff.Write(bnr)
		serial.Flush()
		return txn.Set(b.MsgBlock().Header.PrevBlock.CloneBytes(), buff.Bytes())
	})
}

// GetBlock returnes a *Block looking for the block corresponding to the hash passed
func GetBlock(hash *chainhash.Hash) (*blocks.Block, error) {
	var loadedBlockBytes []byte
	err := instance.View(func(txn *badger.Txn) error {
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
	height, blockBytes := loadedBlockBytes[:4], loadedBlockBytes[4:]
	block, err := btcutil.NewBlockFromBytes(blockBytes)
	if err != nil {
		return nil, err
	}
	h := int32(binary.LittleEndian.Uint32(height))
	block.SetHeight(int32(h))
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
func StoredBlocks() (map[int32]string, error) {
	blocks := make(map[int32]string)
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

			block, err := item.Value()
			if err != nil {
				return err
			}
			height := block[:4]
			h := int32(binary.LittleEndian.Uint32(height))

			blocks[h] = hash.String()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

// GetAll returnes all the blocks stored in badger
func GetAll() ([]blocks.Block, error) {
	var storedBlocks []blocks.Block
	err := instance.View(func(txn *badger.Txn) error {
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

			height, blockBytes := block[:4], block[4:]
			b, err := btcutil.NewBlockFromBytes(blockBytes)
			if err != nil {
				return err
			}
			h := int32(binary.LittleEndian.Uint32(height))
			b.SetHeight(int32(h))
			storedBlocks = append(storedBlocks, blocks.Block{Block: *b})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return storedBlocks, nil
}

// LastBlock returnes the last block stored in the blockchain
func LastBlock() (*chainhash.Hash, error) {
	var loadedBlockBytes []byte
	err := instance.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("last"))
		if err != nil {
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
	hash, err := chainhash.NewHash(loadedBlockBytes)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// DeleteBlock inserts in the db the block as []byte passed
func DeleteBlock(hash *chainhash.Hash) error {
	err := instance.Update(func(txn *badger.Txn) error {
		return txn.Delete(hash.CloneBytes())
	})
	return err
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
