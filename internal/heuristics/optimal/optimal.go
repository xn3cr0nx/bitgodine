package optimal

import (
	"errors"
	"math"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
)

// ChangeOutput returnes the index of the output which value is less than any inputs value, if there is any
func ChangeOutput(tx *dgraph.Transaction) (uint32, error) {
	// max value int64
	var minInput int64 = 9223372036854775807
	for _, in := range tx.Inputs {
		spentTx, err := dgraph.GetTx(in.Hash)
		if err != nil {
			return 0, err
		}
		value := spentTx.Outputs[int(in.Vout)].Value
		minInput = int64(math.Min(float64(minInput), float64(value)))
	}
	var lowerOuts []uint32
	for o, out := range tx.Outputs {
		if out.Value < minInput {
			lowerOuts = append(lowerOuts, uint32(o))
		}
	}
	if len(lowerOuts) > 1 {
		return 0, errors.New("More than an out with lower value of an input, ineffective heuristic")
	}
	if len(lowerOuts) == 1 {
		return lowerOuts[0], nil
	}
	return 0, errors.New("No matching output with value inferior to every input")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(tx *dgraph.Transaction) bool {
	_, err := ChangeOutput(tx)
	if err == nil {
		return true
	}
	return false
}
