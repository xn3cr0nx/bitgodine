package blockchain_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/xn3cr0nx/bitgodine/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine/internal/blocks"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var _ = Describe("Blockchain", func() {
	var (
		bc *Blockchain
	)

	BeforeEach(func() {
		logger.Setup()
	})

	Context("Load blockchain data files", func() {
		It("Should have a Maps with length greater than 0", func() {
			Expect(len(bc.Maps)).NotTo(Equal(0))
		})

		It("Should parse a block correctly out of file 300", func() {
			slice := []uint8(bc.Maps[len(bc.Maps)-1])
			block, err := blocks.Parse(&slice)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(block.Hash()).ToNot(BeNil())
			Expect(block.Hash().String()).ToNot(BeEmpty())
		})

		It("Should consequently parse two data files", func() {
			n := 0
			nBlocks := 0
			for n < 2 {
				slice := []uint8(bc.Maps[n])
				for len(slice) > 0 {
					block, err := blocks.Parse(&slice)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(block.Hash()).ToNot(BeNil())
					Expect(block.Hash().String()).ToNot(BeEmpty())
					nBlocks++
				}
				fmt.Println("nblock after step", n, ":", nBlocks)
				n++
			}
			Expect(nBlocks).Should(BeNumerically(">", 0))
		})
	})
})
