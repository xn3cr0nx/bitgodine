package redis_test

import (
	"fmt"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/spf13/viper"

	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine/internal/blocks"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/test"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing key value storage blocks methods", func() {
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

	Context("Testing storing methods", func() {
		It("Should check block is already stored", func() {
			stored := db.IsStored(chaincfg.MainNetParams.GenesisHash.String())
			Expect(stored).To(BeTrue())
		})

		It("Should fail storing already stored block", func() {
			block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			var txs []models.Tx
			for _, tx := range block.Transactions() {
				txs = append(txs, test.TxToModel(tx, block.Height(), block.Hash().String(), block.MsgBlock().Header.Timestamp))
			}
			err := db.StoreBlock(test.BlockToModel(block), txs)
			Expect(err.Error()).To(Equal(fmt.Sprintf("block %s already exists", chaincfg.MainNetParams.GenesisHash)))
		})

		It("Should correctly store a block", func() {
			blockExample, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
			Expect(err).ToNot(HaveOccurred())
			blockExample.SetHeight(181)
			var txs []models.Tx
			for _, tx := range blockExample.Transactions() {
				txs = append(txs, test.TxToModel(tx, blockExample.Height(), blockExample.Hash().String(), blockExample.MsgBlock().Header.Timestamp))
			}
			err = db.StoreBlock(test.BlockToModel(blockExample), txs)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Testing retrieve methods", func() {
		It("Should correctly fetch genesis block by hash", func() {
			block, err := db.GetBlockFromHash(chaincfg.MainNetParams.GenesisHash.String())
			Expect(err).ToNot(HaveOccurred())
			Expect(block.ID).To(Equal(chaincfg.MainNetParams.GenesisHash.String()))
			Expect(block.Height).To(Equal(int32(0)))
		})

		It("Should correctly fetch genesis block by height", func() {
			block, err := db.GetBlockFromHeight(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(block.ID).To(Equal(chaincfg.MainNetParams.GenesisHash.String()))
			Expect(block.Height).To(Equal(int32(0)))
		})

		// TODO: GetBlockFromHeightRange

		It("Should correctly get last block height", func() {
			blockExample, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
			Expect(err).ToNot(HaveOccurred())
			blockExample.SetHeight(181)
			var txs []models.Tx
			for _, tx := range blockExample.Transactions() {
				txs = append(txs, test.TxToModel(tx, blockExample.Height(), blockExample.Hash().String(), blockExample.MsgBlock().Header.Timestamp))
			}
			err = db.StoreBlock(test.BlockToModel(blockExample), txs)
			Expect(err).ToNot(HaveOccurred())

			height, err := db.GetLastBlockHeight()
			Expect(err).ToNot(HaveOccurred())
			Expect(height).To(Equal(int32(181)))
		})

		It("Should correctly fetch last block", func() {
			block, err := db.LastBlock()
			Expect(err).ToNot(HaveOccurred())
			Expect(block.ID).To(Equal(chaincfg.MainNetParams.GenesisHash.String()))
			Expect(block.Height).To(Equal(int32(0)))
		})

		It("Should correctly fetch list of stored blocks", func() {
			list, err := db.GetStoredBlocks()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(list)).To(Equal(1))
		})

		// TODO: GetBlockTxOutputsFromRange
	})

	Context("Texting remove block", func() {
		It("Should remove stored block by hash", func() {
			err := db.RemoveBlock(&models.Block{ID: chaincfg.MainNetParams.GenesisHash.String()})
			Expect(err).ToNot(HaveOccurred())
			stored := db.IsStored(chaincfg.MainNetParams.GenesisHash.String())
			Expect(stored).To(BeFalse())
		})

		It("Should remove last stored block", func() {
			err := db.RemoveLastBlock()
			Expect(err).ToNot(HaveOccurred())
			stored := db.IsStored(chaincfg.MainNetParams.GenesisHash.String())
			Expect(stored).To(BeFalse())
		})
	})

})
