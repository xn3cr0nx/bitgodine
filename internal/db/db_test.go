package db

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/database"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"gopkg.in/go-playground/assert.v1"

	"github.com/stretchr/testify/suite"
)

type TestDBSuite struct {
	suite.Suite
	level *database.DB
}

func (suite *TestDBSuite) SetupSuite() {
	wd, err := os.Getwd()
	assert.Equal(suite.T(), err, nil)
	conf := &Config{
		Dir:  filepath.Join(wd, "..", ".."),
		Name: "leveldb",
		Net:  wire.MainNet,
	}
	suite.level, _ = Instance(conf)
	assert.NotEqual(suite.T(), suite.level, nil)

	suite.Setup()
}

func (suite *TestDBSuite) Setup() {
	if !IsStored(chaincfg.MainNetParams.GenesisHash) {
		block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
		err := StoreBlock(&blocks.Block{Block: *block})
		assert.Equal(suite.T(), err, nil)
	}
}

func (suite *TestDBSuite) TearDownSuite() {
	(*suite.level).Close()
}

func (suite *TestDBSuite) TestStoreBlock() {
	block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
	err := StoreBlock(&blocks.Block{Block: *block})
	assert.Equal(suite.T(), err.Error(), fmt.Sprintf("block %s already exists", chaincfg.MainNetParams.GenesisHash))
}

func (suite *TestDBSuite) TestGetBlock() {
	block, err := GetBlock(chaincfg.MainNetParams.GenesisHash)
	assert.Equal(suite.T(), err, nil)
	assert.NotEqual(suite.T(), block, nil)
}

func TestDB(t *testing.T) {
	suite.Run(t, new(TestDBSuite))
}
