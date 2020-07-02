package skipped_test

import (
	"fmt"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/blocks"
	"github.com/xn3cr0nx/bitgodine/internal/skipped"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with Ginkgo", func() {
	var (
		s *skipped.Skipped
	)

	BeforeEach(func() {
		logger.Setup()
		var err error
		s = skipped.NewSkipped()
		Expect(s).ToNot(BeNil())

		if !s.IsStored(chaincfg.MainNetParams.GenesisHash) {
			block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
			block.SetHeight(int32(0))
			err = s.StoreBlock(&blocks.Block{Block: *block})
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		viper.SetDefault("dbDir", filepath.Join(".", "test"))
	})

	It("Should fail storing a new block", func() {
		block := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
		block.SetHeight(int32(0))
		err := s.StoreBlock(&blocks.Block{Block: *block})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(fmt.Sprintf("block %s already exists", chaincfg.MainNetParams.GenesisHash)))
	})

	It("Should fetch all stored blocks", func() {
		blocks := s.GetStoredBlocks()
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
})
