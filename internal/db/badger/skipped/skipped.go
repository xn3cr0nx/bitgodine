package skipped

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/badger"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	bdg "github.com/xn3cr0nx/bitgodine_code/internal/db/badger"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Skipped instance of key value store designed to treat block structs
type Skipped struct {
	DB     *badger.DB
	Blocks map[chainhash.Hash]blocks.Block
	memory bool
}

// NewSkipped creates a new instance of Skipped
func NewSkipped(conf *bdg.Config, memory bool) (*Skipped, error) {
	opts := badger.DefaultOptions
	opts.Dir, opts.ValueDir = conf.Dir, conf.Dir
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	b := make(map[chainhash.Hash]blocks.Block, 0)
	return &Skipped{DB: db, Blocks: b, memory: memory}, nil
}

// StoreBlock inserts in the db the block as []byte passed
func (s *Skipped) StoreBlock(b *blocks.Block) error {
	// block validation
	if s.IsStored(b.Hash()) {
		return errors.New(fmt.Sprintf("block %s already exists", b.Hash().String()))
	}
	err := s.DB.Update(func(txn *badger.Txn) error {
		bytes, err := b.Bytes()
		if err != nil {
			return err
		}
		return txn.Set(b.Hash().CloneBytes(), bytes)
	})
	if s.memory {
		s.Blocks[*b.Hash()] = *b
	}
	return err
}

// StoreBlockPrevHash inserts in the db the block as []byte passed, using the previous hash as key
func (s *Skipped) StoreBlockPrevHash(b *blocks.Block) error {
	err := s.DB.Update(func(txn *badger.Txn) error {
		bytes, err := b.Bytes()
		if err != nil {
			return err
		}
		return txn.Set(b.MsgBlock().Header.PrevBlock.CloneBytes(), bytes)
	})
	if s.memory {
		s.Blocks[b.MsgBlock().Header.PrevBlock] = *b
	}
	return err
}

// GetBlock returnes a *Block looking for the block corresponding to the hash passed
func (s *Skipped) GetBlock(hash *chainhash.Hash) (*blocks.Block, error) {
	var loadedBlockBytes []byte
	err := s.DB.View(func(txn *badger.Txn) error {
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
func (s *Skipped) IsStored(hash *chainhash.Hash) bool {
	err := s.DB.View(func(txn *badger.Txn) error {
		_, err := txn.Get(hash.CloneBytes())
		if err != nil {
			return err
		}
		return nil
	})
	return err == nil
}

// StoredBlocks is an utility functions that returnes the list of stored blocks hash
func (s *Skipped) StoredBlocks() ([]string, error) {
	var blocks []string
	err := s.DB.View(func(txn *badger.Txn) error {
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
func (s *Skipped) GetAll() ([]blocks.Block, error) {
	var storedBlocks []blocks.Block
	err := s.DB.View(func(txn *badger.Txn) error {
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
	if s.memory {
		for _, block := range storedBlocks {
			s.Blocks[block.MsgBlock().Header.PrevBlock] = block
		}
	}
	return storedBlocks, nil
}

// DeleteBlock inserts in the db the block as []byte passed
func (s *Skipped) DeleteBlock(hash *chainhash.Hash) error {
	err := s.DB.Update(func(txn *badger.Txn) error {
		return txn.Delete(hash.CloneBytes())
	})
	if s.memory {
		delete(s.Blocks, *hash)
	}
	return err
}
