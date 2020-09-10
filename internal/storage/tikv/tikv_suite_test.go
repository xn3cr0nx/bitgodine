package tikv_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDbblocks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kv Suite")
}
