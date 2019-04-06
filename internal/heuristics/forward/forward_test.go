package forward

import (
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/dgo"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/suite"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
	"gopkg.in/go-playground/assert.v1"
)

type TestForwardSuite struct {
	suite.Suite
	dgraph *dgo.Dgraph
	db     *badger.DB
}

func (suite *TestForwardSuite) SetupSuite() {
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

func (suite *TestForwardSuite) Setup() {
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

func (suite *TestForwardSuite) TearDownSuite() {
	(*suite.db).Close()
}

func (suite *TestForwardSuite) TestChangeOutput() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := block.Transactions()[1]
	t := &txs.Tx{Tx: *testTx}
	vout, err := ChangeOutput(t)
	assert.Equal(suite.T(), err, nil)
	assert.Equal(suite.T(), vout, uint32(1))
}

func (suite *TestForwardSuite) TestVulnerable() {
	// Block 181 is vulnerable because both transaction a16f3ce4dd5deb92d98ef5cf8afeaf0775ebca408f708b2146c4fb42b41e14be
	// in block 181 and transaction 591e91f809d716912ca1d4a9295e70c3e78bab077683f79350f101da64588073 in block 182 use as
	// input a utxo spendable by the address 12cbQLTFMXRnSzktFkuoG3eHoMeFtpTu3S
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := block.Transactions()[1]
	t := &txs.Tx{Tx: *testTx}
	v := Vulnerable(t)
	assert.Equal(suite.T(), v, true)
}

func TestForward(t *testing.T) {
	suite.Run(t, new(TestForwardSuite))
}
