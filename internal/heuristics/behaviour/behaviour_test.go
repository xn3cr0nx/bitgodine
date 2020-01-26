package behaviour_test

import (
	"github.com/btcsuite/btcutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger/kv"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	. "github.com/xn3cr0nx/bitgodine_server/internal/heuristics/behaviour"
	"github.com/xn3cr0nx/bitgodine_server/internal/test"
)

var _ = Describe("Behaviour", func() {
	var (
		db    *kv.KV
		block models.Block
	)

	BeforeEach(func() {
		logger.Setup()

		db, err := test.InitTestDB()
		Expect(err).ToNot(HaveOccurred())
		Expect(db).ToNot(BeNil())

		b, _ := btcutil.NewBlockFromBytes(test.Block181Bytes)
		b.SetHeight(181)
		// Expect(err).ToNot(HaveOccurred())
		block = test.BlockToModel(b)
		_, err = db.StoreBlock(&block)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := test.CleanTestDB(db)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Testing client behaviour heuristic", func() {
		It("should test change output", func() {
			vout, err := ChangeOutput(db, &block.Transactions[1])
			Expect(err).ToNot(HaveOccurred())
			Expect(vout).ToNot(Equal(1))
		})

		It("should test vulnerable", func() {
			vuln := Vulnerable(db, &block.Transactions[1])
			Expect(vuln).ToNot(BeTrue())
		})
	})
})
