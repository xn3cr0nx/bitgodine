package visitor_test

import (
	"github.com/btcsuite/btcutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint/memory"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
)

var _ = Describe("Clusterizer", func() {
	var (
		cltz   visitor.Clusterizer
		set    memory.DisjointSet
		block  blocks.Block
		txItem visitor.TransactionItem
	)

	BeforeEach(func() {
		set = memory.NewDisjointSet()
		cltz = visitor.NewClusterizer(&set)
		blockExample, err := btcutil.NewBlockFromBytes(blocks.Block181Bytes)
		blockExample.SetHeight(181)
		block = blocks.Block{Block: *blockExample}
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should visit block begin and return nil", func() {
		res := cltz.VisitBlockBegin(&block, block.Height())
		Expect(block.Height()).To(Equal(int32(181)))
		Expect(res).To(BeNil())
	})

	It("Should visit transaction begin and return transaction item", func() {
		txItem = cltz.VisitTransactionBegin(nil)
		Expect(txItem).ToNot(BeNil())
	})

	It("Should visit transaction input", func() {
		cltz.VisitTransactionInput(*block.MsgBlock().Transactions[0].TxIn[0], nil, &txItem, visitor.Utxo("test"))
	})

	It("Should visit transaction output", func() {
		utxo, err := cltz.VisitTransactionOutput(*block.MsgBlock().Transactions[0].TxOut[0], nil, &txItem)
		Expect(err).ToNot(HaveOccurred())
		Expect(utxo).ToNot(BeNil())
		Expect(utxo).To(Equal(visitor.Utxo("1JSW4QekxPokWWU4hcRwrheZbZKSkFz9oc")))
	})

	// It("Should visit transaction end", func() {
	// 	cltz.VisitTransactionEnd(txs.Tx{Tx: block.MsgBlock().Transactions[0]}, nil, &txItem)
	// })
})
