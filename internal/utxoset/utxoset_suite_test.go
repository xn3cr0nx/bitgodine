package utxoset_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUtxoset(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utxoset Suite")
}
