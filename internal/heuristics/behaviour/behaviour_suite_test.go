package behaviour_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBehaviour(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Behaviour Suite")
}
