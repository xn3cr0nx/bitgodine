package blocks_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBlocks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blocks Suite")
}
