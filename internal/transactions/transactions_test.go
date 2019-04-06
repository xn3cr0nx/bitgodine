package txs

import (
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/dgo"
	"github.com/mitchellh/go-homedir"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
	"gopkg.in/go-playground/assert.v1"

	"github.com/stretchr/testify/suite"
)

type TestTransactionsSuite struct {
	suite.Suite
	dgraph *dgo.Dgraph
	db     *badger.DB
}

func (suite *TestTransactionsSuite) SetupSuite() {
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

func (suite *TestTransactionsSuite) Setup() {
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

func (suite *TestTransactionsSuite) TearDownSuite() {
	(*suite.db).Close()
}

func (suite *TestTransactionsSuite) TestIsCoinbase() {
	genesis := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
	firstCoinbase := &Tx{Tx: *genesis.Transactions()[0]}
	assert.Equal(suite.T(), firstCoinbase.IsCoinbase(), true)
}

func (suite *TestTransactionsSuite) TestGetSpentTx() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := &Tx{Tx: *block.Transactions()[1]}
	spentTx, err := testTx.GetSpentTx(0)
	assert.Equal(suite.T(), err, nil)
	assert.Equal(suite.T(), spentTx.Hash().String(), "f4184fc596403b9d638783cf57adfe4c75c605f6356fbc91338530e9831e9e16")
	assert.Equal(suite.T(), spentTx.MsgTx().TxIn[0].PreviousOutPoint.Hash.String(), "0437cd7f8525ceed2324359c2d0ba26006d92d856a9c20fa0241106ee5a597c9")
}

func (suite *TestTransactionsSuite) TestIsSpent() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := &Tx{Tx: *block.Transactions()[1]}
	spentTx, err := testTx.GetSpentTx(0)
	assert.Equal(suite.T(), err, nil)
	assert.Equal(suite.T(), spentTx.IsSpent(1), true)
}

func (suite *TestTransactionsSuite) TestGetSpendingTx() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := &Tx{Tx: *block.Transactions()[1]}
	spendingTx, err := testTx.GetSpendingTx(1)
	assert.Equal(suite.T(), err, nil)
	assert.Equal(suite.T(), spendingTx.Hash().String(), "591e91f809d716912ca1d4a9295e70c3e78bab077683f79350f101da64588073")
}

func TestTransactions(t *testing.T) {
	suite.Run(t, new(TestTransactionsSuite))
}
