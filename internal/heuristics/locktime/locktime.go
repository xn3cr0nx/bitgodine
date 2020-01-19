// Package locktime heuristic
// It checks for each output of a transaction, if the spending
// transactions locktime is the same of the original transaction. In this case,
// for the percentage, we just count each transaction that can be coupled with
// a transaction that has the same locktime (if its different from the default
// value).
package locktime

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
)

// ChangeOutput returnes the index of the change output address based on locktime heuristic:
// Bitcoin Core sets the locktime to the current block height to prevent fee sniping.
// If all outputs have been spent, and there is only one output that has been spent
// in a transaction that matches this transaction's locktime behavior, it is the change.
func ChangeOutput(db storage.DB, tx *models.Tx) (uint32, error) {
	locktimeGreaterZero := tx.Locktime > 0
	var candidates []uint32

	for _, out := range tx.Vout {
		// output has been spent, check if locktime is consistent
		if db.IsSpent(tx.TxID, out.Index) {
			spendingTx, err := db.GetFollowingTx(tx.TxID, out.Index)
			if err != nil {
				return 0, err
			}
			if (spendingTx.Locktime > 0) == locktimeGreaterZero {
				candidates = append(candidates, out.Index)
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
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	return err == nil
}
