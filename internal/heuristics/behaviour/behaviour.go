// Package behaviour client heuristic
// This heuristic checks if there are
// output addresses that are the first time they appear in the Blockchain. We
// count the transactions in which appear at least one fresh address in the
// output set.
package behaviour

import (
	"errors"
	"runtime"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	task "github.com/xn3cr0nx/bitgodine_server/internal/errtask"
)

type Worker struct {
	db              storage.DB
	output          models.Output
	vout            int
	outputAddresses []uint32
	blockHeight     int32
}

func (w *Worker) Work() (err error) {
	if w.output.ScriptpubkeyAddress == "" {
		return
	}
	firstOccurence, err := w.db.GetAddressFirstOccurenceHeight(w.output.ScriptpubkeyAddress)
	if err != nil {
		return
	}
	// check if occurence is the first, e.g. the transaction block height is the firstOccurence
	if firstOccurence == w.blockHeight {
		w.outputAddresses = append(w.outputAddresses, uint32(w.vout))
	}
	return nil
}

// ChangeOutput returnes the index of the output which appears for the first time in the chain based on client behaviour heuristic
func ChangeOutput(db storage.DB, tx *models.Tx) (uint32, error) {
	var outputAddresses []uint32
	blockHeight, err := db.GetTxBlockHeight(tx.TxID)
	if err != nil {
		return 0, err
	}

	pool := task.New(runtime.NumCPU() / 2)
	for vout, out := range tx.Vout {
		pool.Do(&Worker{db, out, vout, outputAddresses, blockHeight})
	}
	if err := pool.Shutdown(); err != nil {
		return 0, err
	}

	if len(outputAddresses) > 1 {
		return 0, errors.New("More than an output appear for the first time in the blockchain, ineffective heuristic")
	}
	if len(outputAddresses) == 1 {
		return outputAddresses[0], nil
	}
	return 0, errors.New("No output address for the first time appearing in the blockchain")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	// logger.Error("Client behaviour", err, logger.Params{})
	return err == nil
}
