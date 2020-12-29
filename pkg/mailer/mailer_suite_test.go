package mailer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMailer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mailer Suite")
}
