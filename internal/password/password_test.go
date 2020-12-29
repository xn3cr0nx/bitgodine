package password_test

import (
	. "github.com/xn3cr0nx/bitgodine/internal/password"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Password suite", func() {
	var password string

	BeforeEach(func() {
		password = "password"
	})

	It("Should correctly enrypt the password", func() {
		enc, err := Hash(password)
		Expect(err).ToNot(HaveOccurred())
		Expect(enc).To(HaveLen(60))
	})

	It("Should verify the password is correctly encrypted", func() {
		enc, err := Hash(password)
		Expect(err).ToNot(HaveOccurred())

		b := Verify(enc, "password")
		Expect(b).To(BeTrue())
	})
})
