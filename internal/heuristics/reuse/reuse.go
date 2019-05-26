package reuse

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
)

func contains(recipient []string, element string) bool {
	for _, v := range recipient {
		if v == element {
			return true
		}
	}
	return false
}

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(tx *dgraph.Transaction) (uint32, error) {
	var inputAddresses []string

	for _, in := range tx.Inputs {
		spentTx, err := dgraph.GetTx(in.Hash)
		if err != nil {
			if err.Error() == "Coinbase transaction" {
				continue
			}
			return 0, err
		}
		inputAddresses = append(inputAddresses, spentTx.Outputs[in.Vout].Address)
	}
	// Here on the first matching output, that output is returned as change, but could be a reuse on more outputs?
	for vout, out := range tx.Outputs {
		if contains(inputAddresses, out.Address) {
			return uint32(vout), nil
		}
	}

	return 0, errors.New("No reuse address found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(tx *dgraph.Transaction) bool {
	_, err := ChangeOutput(tx)
	if err == nil {
		return true
	}
	return false
}
