package encoding_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/xn3cr0nx/bitgodine/pkg/encoding"
)

type Mock struct {
	Field string
}

var _ = Describe("Encoding", func() {
	var (
		m Mock
	)

	BeforeEach(func() {
		m = Mock{"testing"}
	})

	Context("Testing storing methods", func() {
		It("Should marshal and unmarshal mocked object", func() {
			res, err := Marshal(m)
			Expect(err).ToNot(HaveOccurred())

			un := Mock{}
			err = Unmarshal(res, &un)
			Expect(err).ToNot(HaveOccurred())

			Expect(un.Field).To(Equal("testing"))
		})
	})
})
