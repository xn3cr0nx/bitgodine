package block_test

import (
	"fmt"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/spf13/viper"

	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/parser/bitcoin"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv/badger"
	"github.com/xn3cr0nx/bitgodine/internal/test"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
)

var _ = Describe("Testing key value storage blocks methods", func() {
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
			service := block.NewService(db, nil)
			err := service.StoreBlock(&model, txs)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		viper.SetDefault("dbDir", filepath.Join(".", "test"))
		err := db.Empty()
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Testing storing methods", func() {
		service := block.NewService(db, nil)

		It("Should check block is already stored", func() {
			stored := db.IsStored(chaincfg.MainNetParams.GenesisHash.String())
			Expect(stored).To(BeTrue())
		})

		It("Should fail storing already stored block", func() {
			blk := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			var txs []tx.Tx
			for _, tx := range blk.Transactions() {
				txs = append(txs, test.TxToModel(tx, blk.Height(), blk.Hash().String(), blk.MsgBlock().Header.Timestamp))
			}
			model := test.BlockToModel(blk)

			err := service.StoreBlock(&model, txs)
			Expect(err.Error()).To(Equal(fmt.Sprintf("block %s already exists", chaincfg.MainNetParams.GenesisHash)))
		})

		It("Should correctly store a block", func() {
			blockExample, err := btcutil.NewBlockFromBytes(bitcoin.Block181Bytes)
			Expect(err).ToNot(HaveOccurred())
			blockExample.SetHeight(181)
			var txs []tx.Tx
			for _, tx := range blockExample.Transactions() {
				txs = append(txs, test.TxToModel(tx, blockExample.Height(), blockExample.Hash().String(), blockExample.MsgBlock().Header.Timestamp))
			}
			model := test.BlockToModel(blockExample)
			err = service.StoreBlock(&model, txs)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Testing retrieve methods", func() {
		service := block.NewService(db, nil)

		It("Should correctly fetch genesis block by hash", func() {
			block, err := service.GetFromHash(chaincfg.MainNetParams.GenesisHash.String())
			Expect(err).ToNot(HaveOccurred())
			Expect(block.ID).To(Equal(chaincfg.MainNetParams.GenesisHash.String()))
			Expect(block.Height).To(Equal(int32(0)))
		})

		It("Should correctly fetch genesis block by height", func() {
			block, err := service.GetFromHeight(0)
			Expect(err).ToNot(HaveOccurred())
			Expect(block.ID).To(Equal(chaincfg.MainNetParams.GenesisHash.String()))
			Expect(block.Height).To(Equal(int32(0)))
		})

		// TODO: GetBlockFromHeightRange

		It("Should correctly get last block height", func() {
			blockExample, err := btcutil.NewBlockFromBytes(bitcoin.Block181Bytes)
			Expect(err).ToNot(HaveOccurred())
			blockExample.SetHeight(181)
			var txs []tx.Tx
			for _, tx := range blockExample.Transactions() {
				txs = append(txs, test.TxToModel(tx, blockExample.Height(), blockExample.Hash().String(), blockExample.MsgBlock().Header.Timestamp))
			}
			model := test.BlockToModel(blockExample)
			err = service.StoreBlock(&model, txs)
			Expect(err).ToNot(HaveOccurred())

			height, err := service.ReadHeight()
			Expect(err).ToNot(HaveOccurred())
			Expect(height).To(Equal(int32(181)))
		})

		It("Should correctly fetch last block", func() {
			block, err := service.GetLast()
			Expect(err).ToNot(HaveOccurred())
			Expect(block.ID).To(Equal(chaincfg.MainNetParams.GenesisHash.String()))
			Expect(block.Height).To(Equal(int32(0)))
		})

		It("Should correctly fetch list of stored blocks", func() {
			list, err := service.GetStored()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(list)).To(Equal(1))
		})

		// TODO: GetBlockTxOutputsFromRange
	})

	Context("Texting remove block", func() {
		service := block.NewService(db, nil)

		It("Should remove stored block by hash", func() {
			err := service.Remove(&block.Block{ID: chaincfg.MainNetParams.GenesisHash.String()})
			Expect(err).ToNot(HaveOccurred())
			stored := db.IsStored(chaincfg.MainNetParams.GenesisHash.String())
			Expect(stored).To(BeFalse())
		})

		It("Should remove last stored block", func() {
			err := service.RemoveLast()
			Expect(err).ToNot(HaveOccurred())
			stored := db.IsStored(chaincfg.MainNetParams.GenesisHash.String())
			Expect(stored).To(BeFalse())
		})
	})

})
