// Package locktime heuristic
// It checks for each output of a transaction, if the spending
// transactions locktime is the same of the original transaction. In this case,
// for the percentage, we just count each transaction that can be coupled with
// a transaction that has the same locktime (if its different from the default
// value).
package locktime

import (
	"fmt"
	"runtime"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	task "github.com/xn3cr0nx/bitgodine_server/internal/errtask"
)

// Worker struct implementing workers pool
type Worker struct {
	db         storage.DB
	txid       string
	vout       uint32
	candidates []uint32
	locktime   uint32
}

// Work executed in the workers pool
func (w *Worker) Work() (err error) {
	fmt.Println("dispatched locktime tx", w.txid, w.vout)
	defer (func() {
		fmt.Println("done locktime tx", w.txid, w.vout)
	})()
	spendingTx, err := w.db.GetFollowingTx(w.txid, w.vout)
	if err != nil {
		return
	}
	if spendingTx.Locktime >= w.locktime {
		w.candidates[w.vout] = 1
	}
	return
}

// ChangeOutput returnes the index of the change output address based on locktime heuristic:
// Bitcoin Core sets the locktime to the current block height to prevent fee sniping.
// If all outputs have been spent, and there is only one output that has been spent
// in a transaction that matches this transaction's locktime behavior, it is the change.
func ChangeOutput(db storage.DB, tx *models.Tx) (c []uint32, err error) {
	if tx.Locktime == 0 {
		return
	}
	candidates := make([]uint32, len(tx.Vout))
	pool := task.New(runtime.NumCPU() / 2)
	for _, out := range tx.Vout {
		pool.Do(&Worker{db, tx.TxID, out.Index, candidates, tx.Locktime})
	}
	if err = pool.Shutdown(); err != nil {
		return
	}

	for i, v := range candidates {
		if v == 1 {
			c = append(c, uint32(i))
		}
	}

	return
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	c, err := ChangeOutput(db, tx)
	return err == nil && len(c) > 0
}
