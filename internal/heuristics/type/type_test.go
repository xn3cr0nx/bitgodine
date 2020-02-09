package class

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger/kv"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_server/internal/test"
)

type TestAddressReuseSuite struct {
	suite.Suite
	db     *kv.KV
	target models.Tx
}

func (suite *TestAddressReuseSuite) SetupSuite() {
	logger.Setup()

	db, err := test.InitTestDB()
	require.Nil(suite.T(), err)
	suite.db = db.(*kv.KV)

	suite.Setup()
}

func (suite *TestAddressReuseSuite) Setup() {
	tx := models.Tx{
		TxID: test.VulnerableFunctions("fbe24c7d93f01bf69f4f7c9d5fd257420ef6a7630c0ade732849958d29e6f9c9"),
		Vin: []models.Input{
			models.Input{
				TxID:         "5c8a9134b185c747cc7ecd9aec35d4358b199412c0db763c2854f33c25f142a2",
				Vout:         1,
				IsCoinbase:   false,
				Scriptsig:    "48304502210085feef77cc739d23fd398d39650fe409791d47f5b7a5b1d02a675bf865e6160802204e8e6a52fd4cd3c2c56a09077a8cc5ed3e84781701b7b16556dc3f8c0820cbb3012102ba35167e50af0626fc432596ebae4e87a0b306f4cb0e568e45a8be4a83aa8aab",
				ScriptsigAsm: "OP_PUSHBYTES_72 304502210085feef77cc739d23fd398d39650fe409791d47f5b7a5b1d02a675bf865e6160802204e8e6a52fd4cd3c2c56a09077a8cc5ed3e84781701b7b16556dc3f8c0820cbb301 OP_PUSHBYTES_33 02ba35167e50af0626fc432596ebae4e87a0b306f4cb0e568e45a8be4a83aa8aab",
			},
		},
		Vout: []models.Output{
			models.Output{
				Index:               0,
				ScriptpubkeyAddress: "33WgwFrQukgEoQ1hXFLvbjqpVH7Bk29Gj2",
				Scriptpubkey:        "a91413fc3de764b9c19f3d53193a9fbeafc7959bb35087",
				ScriptpubkeyAsm:     "OP_HASH160 OP_PUSHBYTES_20 13fc3de764b9c19f3d53193a9fbeafc7959bb350 OP_EQUAL",
				ScriptpubkeyType:    "P2SH",
			},
			models.Output{
				Index:               1,
				ScriptpubkeyAddress: "1FfzzAhMFVFqTEfVgkRTHdjpiwhKy91yas",
				Scriptpubkey:        "76a914a0f1f8b3f9e242b2caf9c4aae7db6bedfdbd608a88ac",
				ScriptpubkeyAsm:     "OP_DUP OP_HASH160 OP_PUSHBYTES_20 a0f1f8b3f9e242b2caf9c4aae7db6bedfdbd608a OP_EQUALVERIFY OP_CHECKSIG",
				ScriptpubkeyType:    "P2PKH",
			},
		},
	}
	suite.target = tx

	spentTx := models.Tx{
		TxID: "5c8a9134b185c747cc7ecd9aec35d4358b199412c0db763c2854f33c25f142a2",
		Vout: []models.Output{
			models.Output{},
			models.Output{
				Index:               1,
				ScriptpubkeyAddress: "1FfzzAhMFVFqTEfVgkRTHdjpiwhKy91yas",
				Scriptpubkey:        "76a914a0f1f8b3f9e242b2caf9c4aae7db6bedfdbd608a88ac",
				ScriptpubkeyAsm:     "OP_DUP OP_HASH160 OP_PUSHBYTES_20 a0f1f8b3f9e242b2caf9c4aae7db6bedfdbd608a OP_EQUALVERIFY OP_CHECKSIG",
				ScriptpubkeyType:    "P2PKH",
			},
		},
	}
	serialized, err := encoding.Marshal(spentTx)
	require.Nil(suite.T(), err)
	err = suite.db.Store("5c8a9134b185c747cc7ecd9aec35d4358b199412c0db763c2854f33c25f142a2", serialized)
	require.Nil(suite.T(), err)
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
