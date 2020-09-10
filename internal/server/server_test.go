package server_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/xn3cr0nx/bitgodine/internal/server"
)

var _ = Describe("Server", func() {
	var (
		port int
	)

	BeforeEach(func() {
		port = 3000
	})

	Context("when a server is created", func() {
		It("a single instance is created", func() {
			s1 := server.NewServer(port, nil, nil, nil)
			s2 := server.NewServer(port+1, nil, nil, nil)
			Expect(s1).To(Equal(s2))
		})
	})

})
