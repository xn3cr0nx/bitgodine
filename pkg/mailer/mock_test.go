package mailer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/xn3cr0nx/bitgodine/pkg/mailer"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var _ = Describe("Mock", func() {
	var client *MockClient

	BeforeEach(func() {
		client = NewMockClient("test")
	})

	It("Should return success for sending email with empty client", func() {
		res, err := client.Send(mail.NewV3Mail())
		Expect(err).ToNot(HaveOccurred())
		Expect(res.StatusCode).To(Equal(200))
	})
})
