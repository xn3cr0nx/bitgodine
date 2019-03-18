package blocks

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
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
	genesisBlockHash := chaincfg.MainNetParams.GenesisHash.String()
	genesisTxHash := "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
	block, err := Parse(&f)
	if err != nil {
		logger.Panic("Test Blockchain", err, logger.Params{"op": "Error reading block"})
	}

	assert.Equal(t, 1, len(block.Transactions()))
	assert.Equal(t, genesisBlockHash, block.Hash().String())
	assert.Equal(t, genesisTxHash, block.Transactions()[0].Hash().String())
}

func TestCheckBlock(t *testing.T) {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")
	block, err := Parse(&f)
	if err != nil {
		logger.Panic("Test Block", err, logger.Params{"op": "Error reading block"})
	}
	assert.Equal(t, true, block.CheckBlock())

	b := Block{}
	assert.Equal(t, false, (&b).CheckBlock())
}

func TestPrintGenesisBlock(t *testing.T) {
	genesisBlock := chaincfg.MainNetParams.GenesisBlock
	coinbase := genesisBlock.Transactions[0].TxIn[0].PreviousOutPoint.Hash.String()
	assert.Equal(t, strings.Repeat("0", 64), coinbase)
	fmt.Println(genesisBlock.Transactions[0].TxIn[0].PreviousOutPoint.Index)
}

func TestCoinbaseValue(t *testing.T) {
	assert.Equal(t, CoinbaseValue(0), int64(5000000000))
	assert.Equal(t, CoinbaseValue(200000), int64(5000000000))
	assert.Equal(t, CoinbaseValue(210000), int64(2500000000))
	assert.Equal(t, CoinbaseValue(420000), int64(1250000000))
	assert.Equal(t, CoinbaseValue(1260000), int64(78125000))
}
