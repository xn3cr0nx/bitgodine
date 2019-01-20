package blocks

import (
	"io/ioutil"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/stretchr/testify/assert"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// func TestBtcutilToBlock(t *testing.T) {
// 	block := btcutil.Block{}
// 	block2 := &Block{}
// 	newBlock := btcutilToBlock(&block)
// 	assert.Equal(t, reflect.TypeOf(block2), reflect.TypeOf(newBlock))
// }

func TestBlockRead(t *testing.T) {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")

	// Tries to read genesis block
	genesisBlockHash := "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"
	genesisTxHash := "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
	block, err := Read(&f)
	if err != nil {
		logger.Panic("Test Blockchain", err, logger.Params{"op": "Error reading block"})
	}

	assert.Equal(t, 1, len(block.Transactions()))
	assert.Equal(t, genesisBlockHash, block.Hash().String())
	assert.Equal(t, genesisTxHash, block.Transactions()[0].Hash().String())
}

func TestCheckBlock(t *testing.T) {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")
	block, err := Read(&f)
	if err != nil {
		logger.Panic("Test Block", err, logger.Params{"op": "Error reading block"})
	}
	assert.Equal(t, true, CheckBlock(block))

	b := btcutil.Block{}
	assert.Equal(t, false, CheckBlock(&b))
}

// func TestWalk(t *testing.T) {

// }
