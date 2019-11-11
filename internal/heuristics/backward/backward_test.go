package backward

import (
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/badger"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/suite"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/dgraph"
	"github.com/xn3cr0nx/bitgodine_server/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_server/internal/db"
	txs "github.com/xn3cr0nx/bitgodine_server/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_server/pkg/logger"
	"gopkg.in/go-playground/assert.v1"
)

type TestBackwardSuite struct {
	suite.Suite
	dgraph *dgraph.Dgraph
	db     *badger.DB
}

func (suite *TestBackwardSuite) SetupSuite() {
	logger.Setup()

	suite.dgraph := dgraph.Instance(dgraph.Conf(), nil)
	suite.dgraph.Setup()

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

func (suite *TestBackwardSuite) Setup() {
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

func (suite *TestBackwardSuite) TearDownSuite() {
	(*suite.db).Close()
}

func (suite *TestBackwardSuite) TestChangeOutput() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := block.Transactions()[1]
	t := &txs.Tx{Tx: *testTx}
	_, err = ChangeOutput(t)
	assert.Equal(suite.T(), err.Error(), "No output address matching backward heurisitic requirements found")
	// assert.Equal(suite.T(), vout, uint32(1))
}

func (suite *TestBackwardSuite) TestVulnerable() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := block.Transactions()[1]
	t := &txs.Tx{Tx: *testTx}
	v := Vulnerable(t)
	assert.Equal(suite.T(), v, false)
}

func TestBackward(t *testing.T) {
	suite.Run(t, new(TestBackwardSuite))
}
