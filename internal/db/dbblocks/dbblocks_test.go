package dbblocks

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/mitchellh/go-homedir"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
	"gopkg.in/go-playground/assert.v1"

	"github.com/stretchr/testify/suite"
)

type TestDbBlocksSuite struct {
	suite.Suite
	db *DbBlocks
}

func (suite *TestDbBlocksSuite) SetupSuite() {
	logger.Setup()

	hd, err := homedir.Dir()
	assert.Equal(suite.T(), err, nil)
	conf := &Config{
		Dir: filepath.Join(hd, ".bitgodine", "badger"),
	}
	suite.db, err = NewDbBlocks(conf)
	assert.Equal(suite.T(), err, nil)
	assert.NotEqual(suite.T(), suite.db, nil)

	suite.Setup()
}

func (suite *TestDbBlocksSuite) Setup() {
	if !suite.db.IsStored(chaincfg.MainNetParams.GenesisHash) {
		block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
		block.SetHeight(int32(0))
		err := suite.db.StoreBlock(&blocks.Block{Block: *block})
		assert.Equal(suite.T(), err, nil)
	}
}

func (suite *TestDbBlocksSuite) TearDownSuite() {
	(*suite.db).Close()
}

func (suite *TestDbBlocksSuite) TestStoreBlock() {
	block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
	block.SetHeight(int32(0))
	err := suite.db.StoreBlock(&blocks.Block{Block: *block})
	// tests that the genesis blocks is already stored (conditions only verified thanks to Setup())
	assert.Equal(suite.T(), err.Error(), fmt.Sprintf("block %s already exists", chaincfg.MainNetParams.GenesisHash))
}

func (suite *TestDbBlocksSuite) TestStoredBlocks() {
	blocks, err := suite.db.StoredBlocks()
	assert.Equal(suite.T(), err, nil)
	genesis, ok := blocks[0]
	assert.Equal(suite.T(), ok, true)
	assert.Equal(suite.T(), genesis, chaincfg.MainNetParams.GenesisHash.String())
}

func (suite *TestDbBlocksSuite) TestGetBlock() {
	block, err := suite.db.GetBlock(chaincfg.MainNetParams.GenesisHash)
	assert.Equal(suite.T(), err, nil)
	assert.Equal(suite.T(), block.Hash().IsEqual(chaincfg.MainNetParams.GenesisHash), true)
	assert.Equal(suite.T(), block.Height(), int32(0))
	assert.Equal(suite.T(), len(block.Transactions()), 1)
}

func TestDbBlocks(t *testing.T) {
	suite.Run(t, new(TestDbBlocksSuite))
}
