package reuse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/dgo"
	"github.com/stretchr/testify/suite"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
	"gopkg.in/go-playground/assert.v1"
)

type TestAddressReuseSuite struct {
	suite.Suite
	dgraph *dgo.Dgraph
	db     *badger.DB
}

func (suite *TestAddressReuseSuite) SetupSuite() {
	logger.Setup()

	DgConf := &dgraph.Config{
		Host: "localhost",
		Port: 9080,
	}
	suite.dgraph = dgraph.Instance(DgConf)
	dgraph.Setup(suite.dgraph)

	wd, err := os.Getwd()
	assert.Equal(suite.T(), err, nil)
	DbConf := &db.Config{
		Dir: filepath.Join(wd, "..", "..", "..", "badger"),
	}
	suite.db, err = db.Instance(DbConf)
	assert.Equal(suite.T(), err, nil)
	assert.NotEqual(suite.T(), suite.db, nil)

	suite.Setup()
}

func (suite *TestAddressReuseSuite) Setup() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)

	if !db.IsStored(block.Hash()) {
		err := db.StoreBlock(&blocks.Block{Block: *block})
		assert.Equal(suite.T(), err, nil)

		for _, tx := range block.Transactions() {
			err := dgraph.StoreTx(tx.Hash().String(), block.Hash().String(), block.Height(), tx.MsgTx().LockTime, tx.MsgTx().TxIn, tx.MsgTx().TxOut)
			assert.Equal(suite.T(), err, nil)
		}
	}
}

func (suite *TestAddressReuseSuite) TearDownSuite() {
	(*suite.db).Close()
}

func (suite *TestAddressReuseSuite) TestChangeOutput() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := block.Transactions()[1]
	t := &txs.Tx{Tx: *testTx}
	vout, err := ChangeOutput(t)
	assert.Equal(suite.T(), err, nil)
	assert.Equal(suite.T(), vout, uint32(1))
}

func (suite *TestAddressReuseSuite) TestVulnerable() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := block.Transactions()[1]
	t := &txs.Tx{Tx: *testTx}
	v := Vulnerable(t)
	assert.Equal(suite.T(), v, true)
}

func TestAddressReuse(t *testing.T) {
	suite.Run(t, new(TestAddressReuseSuite))
}