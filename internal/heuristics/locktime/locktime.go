// Package locktime heuristic
// It checks for each output of a transaction, if the spending
// transactions locktime is the same of the original transaction. In this case,
// for the percentage, we just count each transaction that can be coupled with
// a transaction that has the same locktime (if its different from the default
// value).
package locktime

import (
	"errors"
	"runtime"
	// "fmt"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	task "github.com/xn3cr0nx/bitgodine_server/internal/errtask"
)

type Worker struct {
	db                  storage.DB
	txid                string
	vout                uint32
	candidates          []uint32
	locktimeGreaterZero bool
}

func (w *Worker) Work() (err error) {
	spendingTx, err := w.db.GetFollowingTx(w.txid, w.vout)
	if err != nil {
		return
	}
	if (spendingTx.Locktime > 0) == w.locktimeGreaterZero {
		w.candidates = append(w.candidates, w.vout)
	}
	return
}

// ChangeOutput returnes the index of the change output address based on locktime heuristic:
// Bitcoin Core sets the locktime to the current block height to prevent fee sniping.
// If all outputs have been spent, and there is only one output that has been spent
// in a transaction that matches this transaction's locktime behavior, it is the change.
func ChangeOutput(db storage.DB, tx *models.Tx) (uint32, error) {
	locktimeGreaterZero := tx.Locktime > 0
	var candidates []uint32

	pool := task.New(runtime.NumCPU() / 2)
	for _, out := range tx.Vout {
		pool.Do(&Worker{db, tx.TxID, out.Index, candidates, locktimeGreaterZero})
	}
	if err := pool.Shutdown(); err != nil {
		return 0, err
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
