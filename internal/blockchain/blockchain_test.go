package blockchain

import (
	"io/ioutil"
	"testing"

	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
)

func TestBlockchainRead(t *testing.T) {
	b := Instance(chaincfg.RegressionNetParams)
	b.Read()
	expected := "����"
	assert.Equal(t, expected, string(b.Maps[0][0:10]))

}

func TestBlockRead(t *testing.T) {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")

	// Tries to read genesis block
	genesisBlockHash := "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"
	genesisTxHash := "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
	block, err := ReadBlock(&f)
	if err != nil {
		logger.Panic("Test Blockchain", err, logger.Params{"op": "Error reading block"})
	}

	assert.Equal(t, 1, len(block.Transactions()))
	assert.Equal(t, genesisBlockHash, block.Hash().String())
	assert.Equal(t, genesisTxHash, block.Transactions()[0].Hash().String())
}
