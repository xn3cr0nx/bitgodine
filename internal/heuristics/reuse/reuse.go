// Package reuse heuristic
// This heuristic just checks if an address that appears
// in the input set, appears also in the output set, we just need to count the
// number of transactions in which this condition is satisfied. This happens
// when a user uses the same address to pay and to recollect the exceeding
// amount of a transaction.
package reuse

import (
	"runtime"

	task "github.com/xn3cr0nx/bitgodine_server/internal/errtask"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
)

func contains(recipient []string, element string) bool {
	for _, v := range recipient {
		if v == element {
			return true
		}
	}
	return false
}

// Worker struct implementing workers pool
type Worker struct {
	db             storage.DB
	txid           string
	vout           uint32
	inputAddresses []string
}

// Work executed in the workers pool
func (w *Worker) Work() (err error) {
	spentTx, err := w.db.GetTx(w.txid)
	if err != nil {
		return
	}
	w.inputAddresses[int(w.vout)] = spentTx.Vout[w.vout].ScriptpubkeyAddress
	return
}

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(db storage.DB, tx *models.Tx) (c []uint32, err error) {
	inputAddresses := make([]string, len(tx.Vin))
	pool := task.New(runtime.NumCPU() / 2)
	for _, in := range tx.Vin {
		if in.IsCoinbase {
			continue
		}
		pool.Do(&Worker{db, in.TxID, in.Vout, inputAddresses})
	}
	if err = pool.Shutdown(); err != nil {
		return
	}

	for _, out := range tx.Vout {
		if contains(inputAddresses, out.ScriptpubkeyAddress) {
			c = append(c, out.Index)
		}
	}

	return
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	c, err := ChangeOutput(db, tx)
	return err == nil && len(c) > 0
}
