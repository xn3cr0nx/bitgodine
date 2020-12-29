package reuse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/test"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

type TestAddressReuseSuite struct {
	suite.Suite
	db     kv.DB
	target tx.Tx
}

func (suite *TestAddressReuseSuite) SetupSuite() {
	logger.Setup()

	db, err := test.InitDB()
	require.Nil(suite.T(), err)
	suite.db = db.(kv.DB)

	suite.Setup()
}

func (suite *TestAddressReuseSuite) Setup() {
	// check blockchain is synced at least to block 1000
	h, err := block.ReadHeight(suite.db)
	require.Nil(suite.T(), err)
	require.GreaterOrEqual(suite.T(), h, int32(1000))

	tx, err := tx.GetFromHash(suite.db, nil, test.VulnerableFunctions(("Address Reuse")))
	require.Nil(suite.T(), err)
	suite.target = tx
}

func (suite *TestAddressReuseSuite) TearDownSuite() {
	test.CleanTestDB(suite.db)
}

func (suite *TestAddressReuseSuite) TestChangeOutput() {
	c, err := ChangeOutput(suite.db, nil, &suite.target)
	require.Nil(suite.T(), err)
	assert.Equal(suite.T(), c, []uint32{uint32(1)})
}

func (suite *TestAddressReuseSuite) TestVulnerable() {
	v := Vulnerable(suite.db, nil, &suite.target)
	assert.Equal(suite.T(), v, true)
}

func TestAddressReuse(t *testing.T) {
	suite.Run(t, new(TestAddressReuseSuite))
}
