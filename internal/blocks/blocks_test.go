package blocks

import (
	// "fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/dgo"
	"github.com/stretchr/testify/assert"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/stretchr/testify/suite"
)

type TestBlocksSuite struct {
	suite.Suite
	dgraph *dgo.Dgraph
}

func (suite *TestBlocksSuite) SetupSuite() {
	logger.Setup()
	DgConf := &dgraph.Config{
		Host: "localhost",
		Port: 9080,
	}
	suite.dgraph = dgraph.Instance(DgConf)
	dgraph.Setup(suite.dgraph)

	// suite.Setup()
}

func (suite *TestBlocksSuite) Setup() {
	// block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	// assert.Equal(suite.T(), err, nil)

	// if !db.IsStored(block.Hash()) {
	// 	err := dgraph.StoreBlock(&blocks.Block{Block: *block})
	// 	assert.Equal(suite.T(), err, nil)
	// }
}

func (suite *TestBlocksSuite) TearDownSuite() {
}

func (suite *TestBlocksSuite) TestParse() {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")

	// Tries to read genesis block
	genesisBlockHash := chaincfg.MainNetParams.GenesisHash.String()
	genesisTxHash := "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
	block, err := Parse(&f)
	if err != nil {
		logger.Panic("Test Blockchain", err, logger.Params{"op": "Error reading block"})
	}

	assert.Equal(suite.T(), 1, len(block.Transactions()))
	assert.Equal(suite.T(), genesisBlockHash, block.Hash().String())
	assert.Equal(suite.T(), genesisTxHash, block.Transactions()[0].Hash().String())
}

func (suite *TestBlocksSuite) TestCheckBlock() {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")
	block, err := Parse(&f)
	if err != nil {
		logger.Panic("Test Block", err, logger.Params{"op": "Error reading block"})
	}
	assert.Equal(suite.T(), true, block.CheckBlock())

	b := Block{}
	assert.Equal(suite.T(), false, (&b).CheckBlock())
}

func (suite *TestBlocksSuite) TestPrintGenesisBlock() {
	genesisBlock := chaincfg.MainNetParams.GenesisBlock
	coinbase := genesisBlock.Transactions[0].TxIn[0].PreviousOutPoint.Hash.String()
	assert.Equal(suite.T(), strings.Repeat("0", 64), coinbase)
}

func (suite *TestBlocksSuite) TestCoinbaseValue() {
	assert.Equal(suite.T(), CoinbaseValue(0), int64(5000000000))
	assert.Equal(suite.T(), CoinbaseValue(200000), int64(5000000000))
	assert.Equal(suite.T(), CoinbaseValue(210000), int64(2500000000))
	assert.Equal(suite.T(), CoinbaseValue(420000), int64(1250000000))
	assert.Equal(suite.T(), CoinbaseValue(1260000), int64(78125000))
}

func (suite *TestBlocksSuite) TestGenerateBlock() {
	blockExample, err := btcutil.NewBlockFromBytes(Block181Bytes)
	blockExample.SetHeight(181)
	assert.Equal(suite.T(), err, nil)
	transactions, err := txs.PrepareTransactions(blockExample.Transactions(), blockExample.Height())
	assert.Equal(suite.T(), err, nil)

	block := dgraph.Block{
		Hash:         blockExample.Hash().String(),
		Height:       blockExample.Height(),
		MerkleRoot:   blockExample.MsgBlock().Header.MerkleRoot.String(),
		PrevBlock:    blockExample.MsgBlock().Header.PrevBlock.String(),
		Nonce:        blockExample.MsgBlock().Header.Nonce,
		Time:         blockExample.MsgBlock().Header.Timestamp,
		Version:      blockExample.MsgBlock().Header.Version,
		Bits:         blockExample.MsgBlock().Header.Bits,
		Transactions: transactions,
	}
	genBlock, err := GenerateBlock(&block)
	assert.Equal(suite.T(), err, nil)
	assert.Equal(suite.T(), genBlock.Hash().IsEqual(blockExample.Hash()), true)
}

func TestBlocks(t *testing.T) {
	suite.Run(t, new(TestBlocksSuite))
}
