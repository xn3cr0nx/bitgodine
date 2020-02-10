// Package locktime heuristic
// It checks for each output of a transaction, if the spending
// transactions locktime is the same of the original transaction. In this case,
// for the percentage, we just count each transaction that can be coupled with
// a transaction that has the same locktime (if its different from the default
// value).
package locktime

import (
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"golang.org/x/sync/errgroup"
)

// ChangeOutput returnes the index of the change output address based on locktime heuristic:
// Bitcoin Core sets the locktime to the current block height to prevent fee sniping.
// If all outputs have been spent, and there is only one output that has been spent
// in a transaction that matches this transaction's locktime behavior, it is the change.
func ChangeOutput(db storage.DB, tx *models.Tx) (c []uint32, err error) {
	if tx.Locktime == 0 {
		return
	}

	var g errgroup.Group
	for _, output := range tx.Vout {
		out := output
		g.Go(func() (err error) {
			spendingTx, err := db.GetFollowingTx(tx.TxID, out.Index)
			if err != nil {
				return
			}
			if spendingTx.Locktime >= tx.Locktime {
				c = append(c, out.Index)
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		return
	}

	return
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	c, err := ChangeOutput(db, tx)
	return err == nil && len(c) > 0
}
