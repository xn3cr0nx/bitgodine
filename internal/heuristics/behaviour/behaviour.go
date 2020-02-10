// Package behaviour client heuristic
// This heuristic checks if there are
// output addresses that are the first time they appear in the Blockchain. We
// count the transactions in which appear at least one fresh address in the
// output set.
package behaviour

import (
	"runtime"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	task "github.com/xn3cr0nx/bitgodine_server/internal/errtask"
)

// Worker struct implementing workers pool
type Worker struct {
	db          storage.DB
	output      models.Output
	vout        int
	candidates  []uint32
	blockHeight int32
}

// Work executed in the workers pool
func (w *Worker) Work() (err error) {
	firstOccurence, err := w.db.GetAddressFirstOccurenceHeight(w.output.ScriptpubkeyAddress)
	if err != nil {
		return
	}
	// check if occurence is the first, e.g. the transaction block height is the firstOccurence
	if firstOccurence == w.blockHeight {
		w.candidates = append(w.candidates, uint32(w.vout))
	}
	return
}

// ChangeOutput returnes the index of the output which appears for the first time in the chain based on client behaviour heuristic
func ChangeOutput(db storage.DB, tx *models.Tx) (c []uint32, err error) {
	candidates := make([]uint32, len(tx.Vout))
	blockHeight, err := db.GetTxBlockHeight(tx.TxID)
	if err != nil {
		return
	}

	pool := task.New(runtime.NumCPU(), len(tx.Vout))
	for vout, out := range tx.Vout {
		if out.ScriptpubkeyAddress == "" {
			continue
		}
		pool.Do(&Worker{db, out, vout, candidates, blockHeight})
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
