package visitor_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVisitor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Visitor Suite")
}
