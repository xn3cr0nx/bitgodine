package buffer_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	"github.com/xn3cr0nx/bitgodine/pkg/buffer"
)

var _ = Describe("Testing with Ginkgo", func() {
	It("buffer", func() {

		b := []byte{0xe8, 0x03, 0xd0, 0x07}

		x1, _ := buffer.ReadUint16(&b)

		x2, _ := buffer.ReadUint16(&b)
		assert.Equal(GinkgoT(), "0x03e8 0x07d0", fmt.Sprintf("%#04x %#04x", x1, x2))
	})
})
