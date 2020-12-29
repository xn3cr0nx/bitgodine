package tx_test

import (
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/spf13/viper"

	"github.com/btcsuite/btcutil"
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

var _ = Describe("Testing key value storage transactions methods", func() {
	var (
		db                         kv.DB
		genesisHash, genesisTxHash string
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

		genesisHash = chaincfg.MainNetParams.GenesisHash.String()
		if !db.IsStored(genesisHash) {
			blk := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			genesisTxHash = blk.Transactions()[0].Hash().String()
			blk.SetHeight(int32(0))
			var txs []tx.Tx
			for _, tx := range blk.Transactions() {
				txs = append(txs, test.TxToModel(tx, blk.Height(), blk.Hash().String(), blk.MsgBlock().Header.Timestamp))
			}
			model := test.BlockToModel(blk)
			err := block.StoreBlock(db, &model, txs)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		viper.SetDefault("dbDir", filepath.Join(".", "test"))
		err := db.Empty()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Testing retrieve transactions methods", func() {
		// It("Should check block is already stored", func() {
		// 	stored := db.IsStored(chaincfg.MainNetParams.GenesisHash.String())
		// 	Expect(stored).To(BeTrue())
		// })

		It("Should correctly get transaction from hash", func() {
			tx, err := tx.GetFromHash(db, nil, genesisTxHash)
			Expect(err).ToNot(HaveOccurred())
			Expect(tx.TxID).To(Equal(genesisTxHash))
		})

		It("Should correctly get transaction outputs from hash", func() {
			outputs, err := tx.GetOutputsFromHash(db, nil, genesisTxHash)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(outputs)).To(Equal(1))
		})

		// TODO: GetSpentTxOutput

		// TODO: GetFollowingTx

		It("Should get the list of stored transactions", func() {
			list, err := block.GetStoredTxs(db)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(list)).To(Equal(1))
		})

		It("Should get the height of the block containing transaction's hash", func() {
			height, err := block.GetTxBlockHeight(db, nil, genesisTxHash)
			Expect(err).ToNot(HaveOccurred())
			Expect(height).To(Equal(int32(0)))
		})

		It("Should get the block containing transaction's hash", func() {
			block, err := block.GetTxBlock(db, nil, genesisTxHash)
			Expect(err).ToNot(HaveOccurred())
			Expect(block.ID).To(Equal(genesisHash))
		})

		// TODO: IsSpent
	})
})
