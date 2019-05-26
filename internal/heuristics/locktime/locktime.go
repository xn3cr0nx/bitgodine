package locktime

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
)

// ChangeOutput returnes the index of the change output address based on locktime heuristic:
// Bitcoin Core sets the locktime to the current block height to prevent fee sniping.
// If all outputs have been spent, and there is only one output that has been spent
// in a transaction that matches this transaction's locktime behavior, it is the change.
func ChangeOutput(tx *dgraph.Transaction) (uint32, error) {
	locktimeGreaterZero := tx.Locktime > 0
	var candidates []uint32

	for _, out := range tx.Outputs {
		// output has been spent, check if locktime is consistent
		if dgraph.IsSpent(tx.Hash, out.Vout) {
			// spendingTx, err := tx.GetSpendingTx(uint32(vout))
			spendingTx, err := dgraph.GetFollowingTx(&tx.Hash, &out.Vout)
			if err != nil {
				return 0, err
			}
			if (spendingTx.Locktime > 0) == locktimeGreaterZero {
				candidates = append(candidates, out.Vout)
			}
		} else {
			return 0, errors.New("There is a not spent output, ineffective heuristic")
		}
	}

	if len(candidates) > 1 {
		return 0, errors.New("Many output match the condition for timelock, ineffective heuristic")
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	return 0, errors.New("No output matching the condition for timelock")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(tx *dgraph.Transaction) bool {
	_, err := ChangeOutput(tx)
	if err == nil {
		return true
	}
	return false
}
