package buffer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffer(t *testing.T) {
	b := []byte{0xe8, 0x03, 0xd0, 0x07}
	// First ReadUint16 cut first 2 bytes of b cause it's passed by reference
	x1, _ := ReadUint16(&b)
	// So it would be as pass b[2:] to the second ReadUint16 if it would be passed as copy
	x2, _ := ReadUint16(&b)
	assert.Equal(t, "0x03e8 0x07d0", fmt.Sprintf("%#04x %#04x", x1, x2))
}
