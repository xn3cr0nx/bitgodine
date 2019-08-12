package skipped_test

import (
	"fmt"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/mitchellh/go-homedir"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	database "github.com/xn3cr0nx/bitgodine_code/internal/db/badger"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/badger/skipped"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with Ginkgo", func() {
	var (
		s *skipped.Skipped
	)

	BeforeEach(func() {
		logger.Setup()
		hd, err := homedir.Dir()
		Expect(err).ToNot(HaveOccurred())
		conf := &database.Config{
			Dir: filepath.Join(hd, ".bitgodine", "badger"),
		}
		s, err = skipped.NewSkipped(conf)
		Expect(err).ToNot(HaveOccurred())
		Expect(s).ToNot(BeNil())

		if !s.IsStored(chaincfg.MainNetParams.GenesisHash) {
			block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			block.SetHeight(int32(0))
			err := s.StoreBlock(&blocks.Block{Block: *block})
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("Should fail storing a new block", func() {
		block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
		block.SetHeight(int32(0))
		err := s.StoreBlock(&blocks.Block{Block: *block})
		Expect(err.Error()).To(Equal(fmt.Sprintf("block %s already exists", chaincfg.MainNetParams.GenesisHash)))
	})

	It("Should fetch all stored blocks", func() {
		blocks, err := s.StoredBlocks()
		Expect(err).ToNot(HaveOccurred())
		genesis := blocks[0]
		Expect(genesis).To(Equal(chaincfg.MainNetParams.GenesisHash.String()))
	})

	It("Should fetch a block", func() {
		block, err := s.GetBlock(chaincfg.MainNetParams.GenesisHash)
		Expect(err).ToNot(HaveOccurred())
		Expect(block.Hash().String()).To(Equal(chaincfg.MainNetParams.GenesisHash.String()))
		// Expect(block.Height()).To(Equal(int32(0)))
		Expect(block.Transactions()).To(HaveLen(1))
	})

	AfterEach(func() {
		(*s.Db).Close()
	})

})
