package bitcoin_test

import (
	"io/ioutil"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
	"github.com/xn3cr0nx/bitgodine/internal/parser/bitcoin"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/test"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

type TestBlocksSuite struct {
	suite.Suite
	db storage.DB
}

func (suite *TestBlocksSuite) SetupSuite() {
	logger.Setup()

	db, err := test.InitTestDB()
	Expect(err).ToNot(HaveOccurred())
	Expect(db).ToNot(BeNil())
	suite.db = db
}

func (suite *TestBlocksSuite) Setup() {

}

func (suite *TestBlocksSuite) TearDownSuite() {
}

func (suite *TestBlocksSuite) TestParse() {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")

	genesisBlockHash := chaincfg.MainNetParams.GenesisHash.String()
	genesisTxHash := "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
	block, err := bitcoin.ExtractBlockFromSlice(&f)
	if err != nil {
		logger.Panic("Test Blockchain", err, logger.Params{"op": "Error reading block"})
	}

	assert.Equal(suite.T(), 1, len(block.Transactions()))
	assert.Equal(suite.T(), genesisBlockHash, block.Hash().String())
	assert.Equal(suite.T(), genesisTxHash, block.Transactions()[0].Hash().String())
}

func (suite *TestBlocksSuite) TestCheckBlock() {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")
	block, err := bitcoin.ExtractBlockFromSlice(&f)
	if err != nil {
		logger.Panic("Test Block", err, logger.Params{"op": "Error reading block"})
	}
	assert.Equal(suite.T(), true, block.CheckBlock())

	b := bitcoin.Block{}
	assert.Equal(suite.T(), false, (&b).CheckBlock())
}

func (suite *TestBlocksSuite) TestPrintGenesisBlock() {
	genesisBlock := chaincfg.MainNetParams.GenesisBlock
	coinbase := genesisBlock.Transactions[0].TxIn[0].PreviousOutPoint.Hash.String()
	assert.Equal(suite.T(), strings.Repeat("0", 64), coinbase)
}

func (suite *TestBlocksSuite) TestCoinbaseValue() {
	assert.Equal(suite.T(), bitcoin.CoinbaseValue(0), int64(5000000000))
	assert.Equal(suite.T(), bitcoin.CoinbaseValue(200000), int64(5000000000))
	assert.Equal(suite.T(), bitcoin.CoinbaseValue(210000), int64(2500000000))
	assert.Equal(suite.T(), bitcoin.CoinbaseValue(420000), int64(1250000000))
	assert.Equal(suite.T(), bitcoin.CoinbaseValue(1260000), int64(78125000))
}
