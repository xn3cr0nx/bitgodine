package tikv_test

import (
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/spf13/viper"

	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine/internal/test"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing key value storage transactions methods", func() {
	var (
		db                         storage.DB
		genesisHash, genesisTxHash string
	)

	BeforeEach(func() {
		logger.Setup()
		db, err := test.InitTestDB()
		Expect(err).ToNot(HaveOccurred())
		Expect(db).ToNot(BeNil())

		genesisHash = chaincfg.MainNetParams.GenesisHash.String()
		if !db.IsStored(genesisHash) {
			block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			genesisTxHash = block.Transactions()[0].Hash().String()
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
		// It("Should check block is already stored", func() {
		// 	stored := db.IsStored(chaincfg.MainNetParams.GenesisHash.String())
		// 	Expect(stored).To(BeTrue())
		// })

		It("Should correctly get transaction from hash", func() {
			tx, err := db.GetTx(genesisTxHash)
			Expect(err).ToNot(HaveOccurred())
			Expect(tx.TxID).To(Equal(genesisTxHash))
		})

		It("Should correctly get transaction outputs from hash", func() {
			outputs, err := db.GetTxOutputs(genesisTxHash)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(outputs)).To(Equal(1))
		})

		// TODO: GetSpentTxOutput

		// TODO: GetFollowingTx

		It("Should get the list of stored transactions", func() {
			list, err := db.GetStoredTxs()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(list)).To(Equal(1))
		})

		It("Should get the height of the block containing transaction's hash", func() {
			height, err := db.GetTxBlockHeight(genesisTxHash)
			Expect(err).ToNot(HaveOccurred())
			Expect(height).To(Equal(int32(0)))
		})

		It("Should get the block containing transaction's hash", func() {
			block, err := db.GetTxBlock(genesisTxHash)
			Expect(err).ToNot(HaveOccurred())
			Expect(block.ID).To(Equal(genesisHash))
		})

		// TODO: IsSpent
	})
})
