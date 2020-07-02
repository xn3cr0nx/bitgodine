package locktime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/xn3cr0nx/bitgodine/internal/test"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
	"github.com/xn3cr0nx/bitgodine/pkg/storage"
)

type TestAddressReuseSuite struct {
	suite.Suite
	db     storage.DB
	target models.Tx
}

func (suite *TestAddressReuseSuite) SetupSuite() {
	logger.Setup()

	db, err := test.InitTestDB()
	require.Nil(suite.T(), err)
	suite.db = db.(storage.DB)

	suite.Setup()
}

func (suite *TestAddressReuseSuite) Setup() {
	tx := models.Tx{
		TxID:     test.VulnerableFunctions("Locktime"),
		Locktime: 429990,
		Vin: []models.Input{
			models.Input{
				TxID:         "954860ea2c7176a9e4e34350121f3be9583498f92ed3f478a1af08fe9c118623",
				Vout:         2,
				IsCoinbase:   false,
				Scriptsig:    "47304402207356dc9fda57f3fb560a810b74cebd297fa13a2251516aa02f15e2a5f536600702204cf269158843d00537ccad71608f8f9b73e742d0ed4181e6f6c3b960b11fac4c012103135b36c410dc7a495535ccb4f1304437ff917581404f71492c40f0f50f448d41",
				ScriptsigAsm: "OP_PUSHBYTES_71 304402207356dc9fda57f3fb560a810b74cebd297fa13a2251516aa02f15e2a5f536600702204cf269158843d00537ccad71608f8f9b73e742d0ed4181e6f6c3b960b11fac4c01 OP_PUSHBYTES_33 03135b36c410dc7a495535ccb4f1304437ff917581404f71492c40f0f50f448d41",
			},
		},
		Vout: []models.Output{
			models.Output{
				Index:               0,
				ScriptpubkeyAddress: "1CE4BBFz5FrerkwHuDuE62m84mA3SDfB9P",
				ScriptpubkeyType:    "P2PKH",
			},
			models.Output{
				Index:               1,
				ScriptpubkeyAddress: "18DH45RCQrXaAXbxin7Y8YDB9QPNxe52tj",
				ScriptpubkeyType:    "P2PKH",
			},
			models.Output{
				Index:               2,
				ScriptpubkeyAddress: "19EowvRFJ48dofLEVRokvL5H2hHB6nFgnV",
				ScriptpubkeyType:    "P2PKH",
			},
		},
	}
	suite.target = tx

	spentTx := models.Tx{
		TxID:     "30902fb75974281fe7081b540d08b7f149ee362f4fe4589600477d7491572338",
		Locktime: 429991,
	}
	block := models.Block{ID: "", Height: 0}
	err := suite.db.StoreBlock(block, []models.Tx{tx, spentTx})
	require.Nil(suite.T(), err)

	spentTx = models.Tx{
		TxID:     "e724ed4ae72779fc7ad4a02789143b6c92c3276a38b592a60c51aa0433750aa0",
		Locktime: 430003,
	}
	err = suite.db.StoreBlock(block, []models.Tx{tx, spentTx})
	require.Nil(suite.T(), err)

	spentTx = models.Tx{
		TxID:     "0664032cd4330937e9d28fe2cf614c76494db4d5a3208b37a4957adc3171069a",
		Locktime: 0,
	}
	err = suite.db.StoreBlock(block, []models.Tx{tx, spentTx})
	require.Nil(suite.T(), err)
}

func (suite *TestAddressReuseSuite) TearDownSuite() {
	test.CleanTestDB(suite.db)
}

func (suite *TestAddressReuseSuite) TestChangeOutput() {
	c, err := ChangeOutput(suite.db, &suite.target)
	require.Nil(suite.T(), err)
	assert.Equal(suite.T(), len(c), 2)
}

func (suite *TestAddressReuseSuite) TestVulnerable() {
	v := Vulnerable(suite.db, &suite.target)
	assert.Equal(suite.T(), v, true)
}

func TestAddressReuse(t *testing.T) {
	suite.Run(t, new(TestAddressReuseSuite))
}
