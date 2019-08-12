package skipped_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSkipped(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Skipped Suite")
}
