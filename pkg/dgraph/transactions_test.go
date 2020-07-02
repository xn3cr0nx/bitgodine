package dgraph_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/dgraph"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var _ = Describe("Dgraph transactions", func() {
	var (
		dg *dgraph.Dgraph
	)

	BeforeEach(func() {
		logger.Setup()

		ca, err := cache.NewCache(nil)
		Expect(err).ToNot(HaveOccurred())

		dg = dgraph.Instance(dgraph.Conf(), ca)
		err = dg.Setup()
		Expect(err).ToNot(HaveOccurred())

		_, err = cache.NewCache(nil)
		Expect(err).ToNot(HaveOccurred())

		err = dg.Store(TxMock)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := dg.Empty()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should fetch a transaction", func() {
		tx, err := dg.GetTx(TxMock.TxID)
		Expect(err).ToNot(HaveOccurred())
		Expect(tx.TxID).Should(Equal(TxMock.TxID))
	})

	It("Should fetch the transaction uid", func() {
		uid, err := dg.GetTxUID(TxMock.TxID)
		Expect(err).ToNot(HaveOccurred())
		Expect(uid).ShouldNot(BeNil())
	})

	// It("Should store default coinbase transaction", func() {
	// 	err := dg.StoreCoinbase()
	// 	Expect(err).ToNot(HaveOccurred())

	// 	tx, err := dg.GetTx("0000000000000000000000000000000000000000000000000000000000000000")
	// 	Expect(err).ToNot(HaveOccurred())
	// 	Expect(tx.Vout).Should(HaveLen(3))
	// })

	It("Should get the array of outputs of a transaction", func() {
		outputs, err := dg.GetTxOutputs(TxMock.TxID)
		Expect(err).ToNot(HaveOccurred())
		Expect(outputs).Should(HaveLen(1))
	})

	It("Should get a transaction output", func() {
		output, err := dg.GetSpentTxOutput(TxMock.TxID, 0)
		Expect(err).ToNot(HaveOccurred())
		Expect(output.ScriptpubkeyAddress).Should(Equal(TxMock.Vout[0].ScriptpubkeyAddress))
	})

	// Missing GetFollowingTx

	It("Should get a list of stored transactions hash", func() {
		n := 3
		for i := 0; i < n; i++ {
			err := dg.Store(TxMock)
			Expect(err).ToNot(HaveOccurred())
		}

		txs, err := dg.GetStoredTxs()
		Expect(err).ToNot(HaveOccurred())
		Expect(txs).Should(HaveLen(n + 1))
	})

	// It("Should get the height of a transaction", func() {
	// 	txs, err := dgraph.GetTxBlockHeight(TxMock.TxID)
	// 	Expect(err).ToNot(HaveOccurred())
	// 	Expect(txs).Should(HaveLen(1))
	// })

	// Missing GetTxBlock

	// Missing IsSpent
})
