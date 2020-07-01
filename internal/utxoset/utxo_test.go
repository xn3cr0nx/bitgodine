package utxoset_test

import (
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/xn3cr0nx/bitgodine/internal/utxoset"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var _ = Describe("UtxoSet", func() {
	var (
		hash    string
		utxoset *UtxoSet
	)

	BeforeSuite(func() {
		logger.Setup()

		hash = chaincfg.MainNetParams.GenesisHash.String()
		utxoset = Instance(Conf("test.db", true))
	})

	AfterSuite(func() {
		err := utxoset.Close()
		Expect(err).ToNot(HaveOccurred())
		err = os.Remove("test.db")
		Expect(err).ToNot(HaveOccurred())
	})

	BeforeEach(func() {
		err := utxoset.StoreUtxoSet(hash, map[uint32]string{0: "0xa11", 1: "0xb11"})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := utxoset.DeleteUtxoSet(hash)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should convert a integer number to 8-byte big endian representation", func() {
		b := Itob(3)
		Expect(b).Should(Equal([]byte{0, 0, 0, 0, 0, 0, 0, 3}))
	})

	It("Should convert a 8-byte big endian to its decimal representation", func() {
		v := Btoi([]byte{0, 0, 0, 0, 0, 0, 0, 3})
		Expect(v).Should(Equal(3))
	})

	It("Should get an utxo", func() {
		uid, err := utxoset.GetUtxo(hash, 0)
		Expect(err).ToNot(HaveOccurred())
		Expect(uid).To(Equal("0xa11"))
	})

	It("Should get a stored utxo", func() {
		uid, err := utxoset.GetStoredUtxo(hash, 0)
		Expect(err).ToNot(HaveOccurred())
		Expect(uid).To(Equal("0xa11"))
	})

	It("Should get a utxo set", func() {
		uids, err := utxoset.GetUtxoSet(hash)
		Expect(err).ToNot(HaveOccurred())
		Expect(uids).To(HaveLen(2))
	})

	It("Should store a output set", func() {
		outputs := map[uint32]string{0: "0xa12", 1: "0x2b"}
		err := utxoset.StoreUtxoSet(hash, outputs)
		Expect(err).ToNot(HaveOccurred())

		res, err := utxoset.GetStoredUtxo(hash, 0)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(outputs[0]))
	})

	It("Should delete an utxo set element", func() {
		err := utxoset.DeleteUtxo(hash, 0)
		Expect(err).ToNot(HaveOccurred())

		_, err = utxoset.GetStoredUtxo(hash, 0)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("index out of range"))
	})

	It("Should delete an utxo set", func() {
		err := utxoset.DeleteUtxoSet(hash)
		Expect(err).ToNot(HaveOccurred())

		_, err = utxoset.GetStoredUtxo(hash, 0)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("index out of range"))

		// store utxoset to avoid aftereach error
		err = utxoset.StoreUtxoSet(hash, map[uint32]string{0: "0xa11"})
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should delete the entire utxoset deleting the last element in it", func() {
		err := utxoset.DeleteUtxo(hash, 0)
		Expect(err).ToNot(HaveOccurred())
		err = utxoset.DeleteUtxo(hash, 1)
		Expect(err).ToNot(HaveOccurred())

		_, err = utxoset.GetStoredUtxo(hash, 0)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("index out of range"))

		// store utxoset to avoid aftereach error
		err = utxoset.StoreUtxoSet(hash, map[uint32]string{0: "0xa11"})
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should recover the stored utxo set", func() {
		err := utxoset.Restore(hash)
		Expect(err).ToNot(HaveOccurred())
	})

})
