package power

import (
	"errors"

	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
)

// ChangeOutput returnes the index of the output which value is power of ten, if there is any and only one
func ChangeOutput(tx *txs.Tx) (uint32, error) {
	var powerOutputs []uint32

	for k, out := range tx.MsgTx().TxOut {
		if (out.Value % 10) == 0 {
			powerOutputs = append(powerOutputs, uint32(k))
		}
	}
	if len(powerOutputs) == 0 {
		return 0, errors.New("No output value power of ten found")
	}
	if len(powerOutputs) > 1 {
		return 0, errors.New("More than an output which value is power of ten, heuristic ineffective")
	}
	return powerOutputs[0], nil
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to power heuristic
func Vulnerable(tx *txs.Tx) bool {
	_, err := ChangeOutput(tx)
	if err != nil {
		return false
	}
	return true
}