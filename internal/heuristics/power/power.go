// Package power of ten heuristic
// It checks if in the outputs set there are addresses
// that have an amount in Satoshi that is a power of ten. We count the trans action
// in which this type of output is found.
package power

import (
	"github.com/xn3cr0nx/bitgodine/internal/tx"
)

// PowerOfTen heuristic
type PowerOfTen struct{}

// ChangeOutput returns the index of the output which value is power of ten, if there is any and only one
func (h *PowerOfTen) ChangeOutput(transaction *tx.Tx) (c []uint32, err error) {
	for k, out := range transaction.Vout {
		if (out.Value % 10) == 0 {
			c = append(c, uint32(k))
		}
	}
	return
}

// Vulnerable returns true if the transaction has a privacy vulnerability due to power heuristic
func (h *PowerOfTen) Vulnerable(transaction *tx.Tx) bool {
	c, err := h.ChangeOutput(transaction)
	return err == nil && len(c) > 0
}
