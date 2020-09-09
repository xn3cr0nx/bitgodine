package dgraph_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/xn3cr0nx/bitgodine/internal/storage/dgraph"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var _ = Describe("Testing with Ginkgo", func() {
})

var _ = Describe("Dgraph", func() {
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
	})

	AfterEach(func() {
		err := dg.Empty()
		Expect(err).ToNot(HaveOccurred())
	})

})
