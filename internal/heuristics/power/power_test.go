package power

import (
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/dgo"
	"github.com/stretchr/testify/suite"
	"github.com/xn3cr0nx/bitgodine_server/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	txs "github.com/xn3cr0nx/bitgodine_server/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_server/pkg/logger"
	"gopkg.in/go-playground/assert.v1"
)

type TestPowerSuite struct {
	suite.Suite
	dgraph *dgo.Dgraph
}

func (suite *TestPowerSuite) SetupSuite() {
	logger.Setup()

	DgConf := &dgraph.Config{
		Host: "localhost",
		Port: 9080,
	}
	suite.dgraph = dgraph.Instance(DgConf)
	dgraph.Setup(suite.dgraph)
}

func (suite *TestPowerSuite) TearDownSuite() {
}

func (suite *TestPowerSuite) TestChangeOutput() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := block.Transactions()[1]
	t := &txs.Tx{Tx: *testTx}
	_, err = ChangeOutput(t)
	assert.Equal(suite.T(), err.Error(), "More than an output which value is power of ten, heuristic ineffective")
}

func (suite *TestPowerSuite) TestVulnerable() {
	block, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
	assert.Equal(suite.T(), err, nil)
	testTx := block.Transactions()[1]
	t := &txs.Tx{Tx: *testTx}
	v := Vulnerable(t)
	assert.Equal(suite.T(), v, false)
}

func TestPower(t *testing.T) {
	suite.Run(t, new(TestPowerSuite))
}
