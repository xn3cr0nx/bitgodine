package dgraph_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDgraph(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dgraph Suite")
}
