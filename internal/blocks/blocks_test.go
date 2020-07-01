package blocks_test

import (
	"io/ioutil"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/dgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/xn3cr0nx/bitgodine/internal/blocks"
	txs "github.com/xn3cr0nx/bitgodine/internal/transactions"
	"github.com/xn3cr0nx/bitgodine/pkg/dgraph"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/suite"
)

var _ = Describe("Testing with Ginkgo", func() {
	It("blocks", func() {

		suite.Run(GinkgoT(), new(TestBlocksSuite))
	})
})

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

}

func (suite *TestBlocksSuite) Setup() {

}

func (suite *TestBlocksSuite) TearDownSuite() {
}

func (suite *TestBlocksSuite) TestParse() {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")

	genesisBlockHash := chaincfg.MainNetParams.GenesisHash.String()
	genesisTxHash := "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
	block, err := blocks.Parse(&f)
	if err != nil {
		logger.Panic("Test Blockchain", err, logger.Params{"op": "Error reading block"})
	}

	assert.Equal(suite.T(), 1, len(block.Transactions()))
	assert.Equal(suite.T(), genesisBlockHash, block.Hash().String())
	assert.Equal(suite.T(), genesisTxHash, block.Transactions()[0].Hash().String())
}

func (suite *TestBlocksSuite) TestCheckBlock() {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")
	block, err := blocks.Parse(&f)
	if err != nil {
		logger.Panic("Test Block", err, logger.Params{"op": "Error reading block"})
	}
	assert.Equal(suite.T(), true, block.CheckBlock())

	b := blocks.Block{}
	assert.Equal(suite.T(), false, (&b).CheckBlock())
}

func (suite *TestBlocksSuite) TestPrintGenesisBlock() {
	genesisBlock := chaincfg.MainNetParams.GenesisBlock
	coinbase := genesisBlock.Transactions[0].TxIn[0].PreviousOutPoint.Hash.String()
	assert.Equal(suite.T(), strings.Repeat("0", 64), coinbase)
}

func (suite *TestBlocksSuite) TestCoinbaseValue() {
	assert.Equal(suite.T(), blocks.CoinbaseValue(0), int64(5000000000))
	assert.Equal(suite.T(), blocks.CoinbaseValue(200000), int64(5000000000))
	assert.Equal(suite.T(), blocks.CoinbaseValue(210000), int64(2500000000))
	assert.Equal(suite.T(), blocks.CoinbaseValue(420000), int64(1250000000))
	assert.Equal(suite.T(), blocks.CoinbaseValue(1260000), int64(78125000))
}

func (suite *TestBlocksSuite) TestGenerateBlock() {
	blockExample, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	blockExample.SetHeight(181)
	assert.Equal(suite.T(), err, nil)
	transactions, err := txs.PrepareTransactions(blockExample.Transactions())
	assert.Equal(suite.T(), err, nil)

	block := dgraph.Block{
		Hash:         blockExample.Hash().String(),
		Height:       blockExample.Height(),
		MerkleRoot:   blockExample.MsgBlock().Header.MerkleRoot.String(),
		PrevBlock:    blockExample.MsgBlock().Header.PrevBlock.String(),
		Nonce:        blockExample.MsgBlock().Header.Nonce,
		Timestamp:    blockExample.MsgBlock().Header.Timestamp,
		Version:      blockExample.MsgBlock().Header.Version,
		Bits:         blockExample.MsgBlock().Header.Bits,
		Transactions: transactions,
	}
	genBlock, err := blocks.GenerateBlock(&block)
	assert.Equal(suite.T(), err, nil)
	assert.Equal(suite.T(), genBlock.Hash().IsEqual(blockExample.Hash()), true)
}
