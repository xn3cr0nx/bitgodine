package class

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

	db, err := test.InitTestDB()
	require.Nil(suite.T(), err)
	suite.db = db.(kv.DB)

	suite.Setup()
}

func (suite *TestAddressReuseSuite) Setup() {
	transaction := tx.Tx{
		TxID: test.VulnerableFunctions("Address Type"),
		Vin: []tx.Input{
			{
				TxID:         "5c8a9134b185c747cc7ecd9aec35d4358b199412c0db763c2854f33c25f142a2",
				Vout:         1,
				IsCoinbase:   false,
				Scriptsig:    "48304502210085feef77cc739d23fd398d39650fe409791d47f5b7a5b1d02a675bf865e6160802204e8e6a52fd4cd3c2c56a09077a8cc5ed3e84781701b7b16556dc3f8c0820cbb3012102ba35167e50af0626fc432596ebae4e87a0b306f4cb0e568e45a8be4a83aa8aab",
				ScriptsigAsm: "OP_PUSHBYTES_72 304502210085feef77cc739d23fd398d39650fe409791d47f5b7a5b1d02a675bf865e6160802204e8e6a52fd4cd3c2c56a09077a8cc5ed3e84781701b7b16556dc3f8c0820cbb301 OP_PUSHBYTES_33 02ba35167e50af0626fc432596ebae4e87a0b306f4cb0e568e45a8be4a83aa8aab",
			},
		},
		Vout: []tx.Output{
			{
				Index:               0,
				ScriptpubkeyAddress: "33WgwFrQukgEoQ1hXFLvbjqpVH7Bk29Gj2",
				Scriptpubkey:        "a91413fc3de764b9c19f3d53193a9fbeafc7959bb35087",
				ScriptpubkeyAsm:     "OP_HASH160 OP_PUSHBYTES_20 13fc3de764b9c19f3d53193a9fbeafc7959bb350 OP_EQUAL",
				ScriptpubkeyType:    "P2SH",
			},
			{
				Index:               1,
				ScriptpubkeyAddress: "1FfzzAhMFVFqTEfVgkRTHdjpiwhKy91yas",
				Scriptpubkey:        "76a914a0f1f8b3f9e242b2caf9c4aae7db6bedfdbd608a88ac",
				ScriptpubkeyAsm:     "OP_DUP OP_HASH160 OP_PUSHBYTES_20 a0f1f8b3f9e242b2caf9c4aae7db6bedfdbd608a OP_EQUALVERIFY OP_CHECKSIG",
				ScriptpubkeyType:    "P2PKH",
			},
		},
	}
	suite.target = transaction

	spentTx := tx.Tx{
		TxID: "5c8a9134b185c747cc7ecd9aec35d4358b199412c0db763c2854f33c25f142a2",
		Vout: []tx.Output{
			{},
			{
				Index:               1,
				ScriptpubkeyAddress: "1FfzzAhMFVFqTEfVgkRTHdjpiwhKy91yas",
				Scriptpubkey:        "76a914a0f1f8b3f9e242b2caf9c4aae7db6bedfdbd608a88ac",
				ScriptpubkeyAsm:     "OP_DUP OP_HASH160 OP_PUSHBYTES_20 a0f1f8b3f9e242b2caf9c4aae7db6bedfdbd608a OP_EQUALVERIFY OP_CHECKSIG",
				ScriptpubkeyType:    "P2PKH",
			},
		},
	}

	blk := block.Block{ID: "", Height: 0}
	err := block.StoreBlock(suite.db, &blk, []tx.Tx{transaction, spentTx})
	require.Nil(suite.T(), err)
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
