package skipped

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/xn3cr0nx/bitgodine/internal/blocks"
	"github.com/xn3cr0nx/bitgodine/pkg/badger"
)

// Skipped instance of key value store designed to treat block structs
type Skipped struct {
	blocks map[chainhash.Hash]blocks.Block
}

// NewSkipped creates a new instance of Skipped
func NewSkipped(conf *badger.Config) (s *Skipped) {
	b := make(map[chainhash.Hash]blocks.Block)
	s = &Skipped{b}
	return
}

func (s *Skipped) Len() int {
	return len(s.blocks)
}

// StoreBlock inserts in the db the block as []byte passed
func (s *Skipped) StoreBlock(v interface{}) (err error) {
	b := v.(*blocks.Block)
	// block validation
	if s.IsStored(b.Hash()) {
		err = errors.New("block " + b.Hash().String() + " already exists")
		return
	}
	s.blocks[*b.Hash()] = *b
	return
}

// StoreBlockPrevHash inserts in the db the block as []byte passed, using the previous hash as key
func (s *Skipped) StoreBlockPrevHash(b *blocks.Block) {
	s.blocks[b.MsgBlock().Header.PrevBlock] = *b
}

// GetBlock returnes a *Block looking for the block corresponding to the hash passed
func (s *Skipped) GetBlock(hash *chainhash.Hash) (block blocks.Block, err error) {
	block, ok := s.blocks[*hash]
	if !ok {
		err = errors.New("Block not found")
	}
	return
}

// IsStored returns true if the block corresponding to passed hash is stored in db
func (s *Skipped) IsStored(hash *chainhash.Hash) bool {
	_, ok := s.blocks[*hash]
	return ok
}

// GetStoredBlocks is an utility functions that returnes the list of stored blocks hash
func (s *Skipped) GetStoredBlocks() (blocks []string) {
	for _, b := range s.blocks {
		blocks = append(blocks, b.Hash().String())
	}
	return
}

// DeleteBlock inserts in the db the block as []byte passed
func (s *Skipped) DeleteBlock(hash *chainhash.Hash) {
	delete(s.blocks, *hash)
}

// Empty set blocks map to empty map
func (s *Skipped) Empty() {
	s.blocks = make(map[chainhash.Hash]blocks.Block)
}
