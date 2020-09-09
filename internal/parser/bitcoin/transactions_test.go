package bitcoin_test

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"

	"github.com/xn3cr0nx/bitgodine/internal/parser/bitcoin"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transactions", func() {
	BeforeEach(func() {
		logger.Setup()
	})

	It("Should check if transaction is coinbase", func() {
		genesis := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
		firstCoinbase := &bitcoin.Tx{Tx: *genesis.Transactions()[0]}
		Expect(firstCoinbase.IsCoinbase()).To(BeTrue())
	})
})
