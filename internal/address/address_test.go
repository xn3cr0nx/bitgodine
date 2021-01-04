package address_test

import (
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/spf13/viper"

	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine/internal/address"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv/badger"
	"github.com/xn3cr0nx/bitgodine/internal/test"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing key value storage addresses methods", func() {
	var (
		db kv.DB
	)

	BeforeEach(func() {
		logger.Setup()
		conf := &badger.Config{
			Dir: filepath.Join(".", "test"),
		}
		ca, err := cache.NewCache(nil)
		Expect(err).ToNot(HaveOccurred())
		bdg, err := badger.NewBadger(conf, false)
		Expect(err).ToNot(HaveOccurred())
		db, err = badger.NewKV(bdg, ca)
		Expect(err).ToNot(HaveOccurred())
		Expect(db).ToNot(BeNil())

		if !db.IsStored(chaincfg.MainNetParams.GenesisHash.String()) {
			blk := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			blk.SetHeight(int32(0))
			var txs []tx.Tx
			for _, tx := range blk.Transactions() {
				txs = append(txs, test.TxToModel(tx, blk.Height(), blk.Hash().String(), blk.MsgBlock().Header.Timestamp))
			}
			model := test.BlockToModel(blk)
			blockService := block.NewService(db, nil)
			err := blockService.StoreBlock(&model, txs)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		viper.SetDefault("dbDir", filepath.Join(".", "test"))
		err := db.Empty()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Testing retrieve transactions methods", func() {
		service := address.NewService(db, nil)

		It("Should get address height occurence", func() {
			block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			_, addr, _, err := txscript.ExtractPkScriptAddrs(block.Transactions()[0].MsgTx().TxOut[0].PkScript, &chaincfg.MainNetParams)
			Expect(err).ToNot(HaveOccurred())
			occurences, err := service.GetOccurences(addr[0].String())
			Expect(err).ToNot(HaveOccurred())
			Expect(len(occurences)).To(Equal(1))
		})

		It("Should get address height occurence", func() {
			block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			_, addr, _, err := txscript.ExtractPkScriptAddrs(block.Transactions()[0].MsgTx().TxOut[0].PkScript, &chaincfg.MainNetParams)
			Expect(err).ToNot(HaveOccurred())
			height, err := service.GetFirstOccurenceHeight(addr[0].String())
			Expect(err).ToNot(HaveOccurred())
			Expect(height).To(Equal(int32(0)))
		})

	})
})
