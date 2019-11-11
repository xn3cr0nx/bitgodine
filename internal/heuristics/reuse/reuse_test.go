package reuse

import (
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/dgo"
	"github.com/stretchr/testify/suite"
	"github.com/xn3cr0nx/bitgodine_server/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/dgraph"
	txs "github.com/xn3cr0nx/bitgodine_server/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_server/pkg/logger"
	"gopkg.in/go-playground/assert.v1"
)

type TestAddressReuseSuite struct {
	suite.Suite
	dgraph *dgo.Dgraph
}

func (suite *TestAddressReuseSuite) SetupSuite() {
	logger.Setup()

	DgConf := &dgraph.Config{
		Host: "localhost",
		Port: 9080,
	}
	suite.dgraph = dgraph.Instance(DgConf)
	dgraph.Setup(suite.dgraph)
	// suite.Setup()
}

func (suite *TestAddressReuseSuite) Setup() {
	// block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	// assert.Equal(suite.T(), err, nil)

	// if !db.IsStored(block.Hash()) {
	// 	err := db.StoreBlock(&blocks.Block{Block: *block})
	// 	assert.Equal(suite.T(), err, nil)

	// 	for _, tx := range block.Transactions() {
	// 		err := dgraph.StoreTx(tx.Hash().String(), block.Hash().String(), block.Height(), tx.MsgTx().LockTime, tx.MsgTx().TxIn, tx.MsgTx().TxOut)
	// 		assert.Equal(suite.T(), err, nil)
	// 	}
	// }
}

func (suite *TestAddressReuseSuite) TearDownSuite() {
	// (*suite.db).Close()
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
