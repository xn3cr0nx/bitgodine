package txs_test

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"

	"github.com/dgraph-io/dgo/v2"
	"github.com/xn3cr0nx/bitgodine/internal/mocks"
	"github.com/xn3cr0nx/bitgodine/pkg/dgraph"
	"github.com/xn3cr0nx/bitgodine/pkg/models"

	txs "github.com/xn3cr0nx/bitgodine/internal/transactions"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transactions", func() {
	var (
		dg *dgo.Dgraph
	)

	BeforeEach(func() {
		logger.Setup()

		DgConf := &dgraph.Config{
			Host: "localhost",
			Port: 9080,
		}
		dg = dgraph.Instance(DgConf)
		dgraph.Setup(dg)
	})

	It("Should check if transaction is coinbase", func() {
		genesis := btcutil.NewBlock(chaincfg.MainNetParams.GenesisBlock)
		firstCoinbase := &txs.Tx{Tx: *genesis.Transactions()[0]}
		Expect(firstCoinbase.IsCoinbase()).To(BeTrue())
	})

	It("Should generate a transaction from dgraph Transaction node", func() {
		block, err := btcutil.NewBlockFromBytes(mocks.Block181Bytes)
		Expect(err).ToNot(HaveOccurred())
		txExample := block.Transactions()[1]

		var inputs []models.Input
		for _, in := range txExample.MsgTx().TxIn {
			var txWitness []string
			for _, w := range in.Witness {
				txWitness = append(txWitness, string(w))
			}
			inputs = append(inputs, models.Input{TxID: in.PreviousOutPoint.Hash.String(), Vout: in.PreviousOutPoint.Index, Scriptsig: string(in.SignatureScript), Witness: txWitness})
		}
		var outputs []models.Output
		for _, out := range txExample.MsgTx().TxOut {
			outputs = append(outputs, models.Output{Value: out.Value, Scriptpubkey: string(out.PkScript)})
		}
		transaction := models.Tx{
			TxID:     txExample.Hash().String(),
			Locktime: txExample.MsgTx().LockTime,
			Version:  txExample.MsgTx().Version,
			Vin:      inputs,
			Vout:     outputs,
		}
		genTx, err := txs.GenerateTransaction(&transaction)
		Expect(err).ToNot(HaveOccurred())
		Expect(genTx.Hash().String()).To(Equal(txExample.Hash().String()))
		// assert.Equal(suite.T(), genTx.Hash().IsEqual(txExample.Hash()), true)

	})

	It("Should generate transaction after fetching it", func() {
		tx, err := dgraph.GetTx("9552674e9c19536d69dcf45ccf7ec548939c7cc257581edbc85bc5cd9528cf78")
		Expect(err).NotTo(HaveOccurred())
		transaction, err := txs.GenerateTransaction(&tx)
		Expect(err).NotTo(HaveOccurred())
		Expect(transaction).ToNot(BeNil())
	})
})
