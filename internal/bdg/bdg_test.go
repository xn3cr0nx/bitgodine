package bdg

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/badger"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
	"gopkg.in/go-playground/assert.v1"

	"github.com/stretchr/testify/suite"
)

type TestBadgerSuite struct {
	suite.Suite
	db *badger.DB
}

func contains(recipient []string, element string) bool {
	for _, v := range recipient {
		if v == element {
			return true
		}
	}
	return false
}

func (suite *TestBadgerSuite) SetupSuite() {
	logger.Setup()

	wd, err := os.Getwd()
	assert.Equal(suite.T(), err, nil)
	conf := &Config{
		Dir: filepath.Join(wd, "..", "..", "badger"),
	}
	suite.db, err = Instance(conf)
	assert.Equal(suite.T(), err, nil)
	assert.NotEqual(suite.T(), suite.db, nil)

	suite.Setup()
}

func (suite *TestBadgerSuite) Setup() {
	if !IsStored(chaincfg.MainNetParams.GenesisHash) {
		block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
		err := StoreBlock(&blocks.Block{Block: *block})
		assert.Equal(suite.T(), err, nil)
	}
}

func (suite *TestBadgerSuite) TearDownSuite() {
	(*suite.db).Close()
}

func (suite *TestBadgerSuite) TestStoreBlock() {
	block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
	err := StoreBlock(&blocks.Block{Block: *block})
	// tests that the genesis blocks is already stored (conditions only verified thanks to Setup())
	assert.Equal(suite.T(), err.Error(), fmt.Sprintf("block %s already exists", chaincfg.MainNetParams.GenesisHash))
}

func (suite *TestBadgerSuite) TestStoredBlocks() {
	blocks, err := StoredBlocks()
	assert.Equal(suite.T(), err, nil)
	assert.Equal(suite.T(), contains(blocks, chaincfg.MainNetParams.GenesisHash.String()), true)
}

func (suite *TestBadgerSuite) TestGetBlock() {
	block, err := GetBlock(chaincfg.MainNetParams.GenesisHash)
	assert.Equal(suite.T(), err, nil)
	assert.Equal(suite.T(), block.Hash().IsEqual(chaincfg.MainNetParams.GenesisHash), true)
	assert.Equal(suite.T(), len(block.Transactions()), 1)
}

func TestBadger(t *testing.T) {
	suite.Run(t, new(TestBadgerSuite))
}
