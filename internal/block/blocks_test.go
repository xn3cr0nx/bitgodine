package block_test

import (
	"fmt"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/spf13/viper"

	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/parser/bitcoin"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/storage/badger"
	"github.com/xn3cr0nx/bitgodine/internal/test"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing key value storage blocks methods", func() {
	var (
		db storage.DB
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
			err := block.StoreBlock(db, test.BlockToModel(blk), txs)
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
			blk := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			var txs []tx.Tx
			for _, tx := range blk.Transactions() {
				txs = append(txs, test.TxToModel(tx, blk.Height(), blk.Hash().String(), blk.MsgBlock().Header.Timestamp))
			}
			err := block.StoreBlock(db, test.BlockToModel(blk), txs)
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
			err = block.StoreBlock(db, test.BlockToModel(blockExample), txs)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Testing retrieve methods", func() {
		It("Should correctly fetch genesis block by hash", func() {
			block, err := block.GetFromHash(db, nil, chaincfg.MainNetParams.GenesisHash.String())
			Expect(err).ToNot(HaveOccurred())
			Expect(block.ID).To(Equal(chaincfg.MainNetParams.GenesisHash.String()))
			Expect(block.Height).To(Equal(int32(0)))
		})

		It("Should correctly fetch genesis block by height", func() {
			block, err := block.GetFromHeight(db, nil, 0)
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
			err = block.StoreBlock(db, test.BlockToModel(blockExample), txs)
			Expect(err).ToNot(HaveOccurred())

			height, err := block.ReadHeight(db)
			Expect(err).ToNot(HaveOccurred())
			Expect(height).To(Equal(int32(181)))
		})

		It("Should correctly fetch last block", func() {
			block, err := block.GetLast(db, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(block.ID).To(Equal(chaincfg.MainNetParams.GenesisHash.String()))
			Expect(block.Height).To(Equal(int32(0)))
		})

		It("Should correctly fetch list of stored blocks", func() {
			list, err := block.GetStored(db)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(list)).To(Equal(1))
		})

		// TODO: GetBlockTxOutputsFromRange
	})

	Context("Texting remove block", func() {
		It("Should remove stored block by hash", func() {
			err := block.Remove(db, &block.Block{ID: chaincfg.MainNetParams.GenesisHash.String()})
			Expect(err).ToNot(HaveOccurred())
			stored := db.IsStored(chaincfg.MainNetParams.GenesisHash.String())
			Expect(stored).To(BeFalse())
		})

		It("Should remove last stored block", func() {
			err := block.RemoveLast(db, nil)
			Expect(err).ToNot(HaveOccurred())
			stored := db.IsStored(chaincfg.MainNetParams.GenesisHash.String())
			Expect(stored).To(BeFalse())
		})
	})

})
