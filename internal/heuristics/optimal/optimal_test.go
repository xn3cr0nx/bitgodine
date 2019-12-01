package optimal

import (
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/dgo"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/suite"
	"github.com/xn3cr0nx/bitgodine_server/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_server/internal/db"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	txs "github.com/xn3cr0nx/bitgodine_server/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_server/pkg/logger"
	"gopkg.in/go-playground/assert.v1"
)

type TestOptimalChangeSuite struct {
	suite.Suite
	dgraph *dgo.Dgraph
	db     *badger.DB
}

func (suite *TestOptimalChangeSuite) SetupSuite() {
	logger.Setup()

	DgConf := &dgraph.Config{
		Host: "localhost",
		Port: 9080,
	}
	suite.dgraph = dgraph.Instance(DgConf)
	dgraph.Setup(suite.dgraph)

	hd, err := homedir.Dir()
	assert.Equal(suite.T(), err, nil)
	DbConf := &db.Config{
		Dir: filepath.Join(hd, ".bitgodine", "badger"),
	}
	suite.db, err = db.Instance(DbConf)
	assert.Equal(suite.T(), err, nil)
	assert.NotEqual(suite.T(), suite.db, nil)

	suite.Setup()
}

func (suite *TestOptimalChangeSuite) Setup() {
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

func (suite *TestOptimalChangeSuite) TearDownSuite() {
	(*suite.db).Close()
}

func (suite *TestOptimalChangeSuite) TestChangeOutput() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := block.Transactions()[1]
	t := &txs.Tx{Tx: *testTx}
	_, err = ChangeOutput(t)
	assert.Equal(suite.T(), err.Error(), "More than an out with lower value of an input, ineffective heuristic")
	// assert.Equal(suite.T(), vout, uint32(1))
}

func (suite *TestOptimalChangeSuite) TestVulnerable() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := block.Transactions()[1]
	t := &txs.Tx{Tx: *testTx}
	v := Vulnerable(t)
	assert.Equal(suite.T(), v, false)
}

func TestOptimalChange(t *testing.T) {
	suite.Run(t, new(TestOptimalChangeSuite))
}
