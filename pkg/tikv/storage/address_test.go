package storage_test

import (
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/spf13/viper"

	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine/internal/test"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
	"github.com/xn3cr0nx/bitgodine/pkg/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing key value storage addresses methods", func() {
	var (
		db storage.DB
	)

	BeforeEach(func() {
		logger.Setup()
		db, err := test.InitTestDB()
		Expect(err).ToNot(HaveOccurred())
		Expect(db).ToNot(BeNil())

		if !db.IsStored(chaincfg.MainNetParams.GenesisHash.String()) {
			block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			block.SetHeight(int32(0))
			var txs []models.Tx
			for _, tx := range block.Transactions() {
				txs = append(txs, test.TxToModel(tx, block.Height(), block.Hash().String(), block.MsgBlock().Header.Timestamp))
			}
			err := db.StoreBlock(test.BlockToModel(block), txs)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		viper.SetDefault("dbDir", filepath.Join(".", "test"))
		err := db.Empty()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Testing retrieve transactions methods", func() {
		It("Should get address height occurence", func() {
			block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			_, addr, _, err := txscript.ExtractPkScriptAddrs(block.Transactions()[0].MsgTx().TxOut[0].PkScript, &chaincfg.MainNetParams)
			Expect(err).ToNot(HaveOccurred())
			occurences, err := db.GetAddressOccurences(addr[0].String())
			Expect(err).ToNot(HaveOccurred())
			Expect(len(occurences)).To(Equal(1))
		})

		It("Should get address height occurence", func() {
			block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			_, addr, _, err := txscript.ExtractPkScriptAddrs(block.Transactions()[0].MsgTx().TxOut[0].PkScript, &chaincfg.MainNetParams)
			Expect(err).ToNot(HaveOccurred())
			height, err := db.GetAddressFirstOccurenceHeight(addr[0].String())
			Expect(err).ToNot(HaveOccurred())
			Expect(height).To(Equal(int32(0)))
		})

	})
})
