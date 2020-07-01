package reuse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/xn3cr0nx/bitgodine/internal/test"
	"github.com/xn3cr0nx/bitgodine/pkg/badger/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

type TestAddressReuseSuite struct {
	suite.Suite
	db     *kv.KV
	target models.Tx
}

func (suite *TestAddressReuseSuite) SetupSuite() {
	logger.Setup()

	db, err := test.InitDB()
	require.Nil(suite.T(), err)
	suite.db = db.(*kv.KV)

	suite.Setup()
}

func (suite *TestAddressReuseSuite) Setup() {
	// check blockchain is synced at least to block 1000
	h, err := suite.db.GetLastBlockHeight()
	require.Nil(suite.T(), err)
	require.GreaterOrEqual(suite.T(), h, int32(1000))

	tx, err := suite.db.GetTx(test.VulnerableFunctions(("Address Reuse")))
	require.Nil(suite.T(), err)
	suite.target = tx
}

func (suite *TestAddressReuseSuite) TearDownSuite() {
	test.CleanTestDB(suite.db)
}

func (suite *TestAddressReuseSuite) TestChangeOutput() {
	c, err := ChangeOutput(suite.db, &suite.target)
	require.Nil(suite.T(), err)
	assert.Equal(suite.T(), c, []uint32{uint32(1)})
}

func (suite *TestAddressReuseSuite) TestVulnerable() {
	v := Vulnerable(suite.db, &suite.target)
	assert.Equal(suite.T(), v, true)
}

func TestAddressReuse(t *testing.T) {
	suite.Run(t, new(TestAddressReuseSuite))
}
