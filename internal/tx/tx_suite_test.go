package tx_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tx Suite")
}
