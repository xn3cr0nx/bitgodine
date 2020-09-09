// Package forward heuristic
// It checks the transactions that come
// after the one in which we want to find the change address.
package forward

import (
	"errors"

	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(db storage.DB, tx *models.Tx) (c []uint32, err error) {
	var inputAddresses []string

	logger.Debug("Forward Heuristic", "transaction "+tx.TxID, logger.Params{})

	for _, in := range tx.Vin {
		if in.IsCoinbase {
			continue
		}
		spentTx, e := db.GetTx(in.TxID)
		if e != nil {
			return nil, e
		}
		inputAddresses = append(inputAddresses, spentTx.Vout[in.Vout].ScriptpubkeyAddress)
	}

	for _, out := range tx.Vout {
		spendingTx, e := db.GetFollowingTx(tx.TxID, out.Index)
		if e != nil {
			// transaction not found => output not yet spent, but we can identify the change output anyway
			if e.Error() == "transaction not found" {
				continue
			}
			return nil, e
		}
		index := out.Index
		for _, spendingIn := range spendingTx.Vin {
			// check if the input is the one the spending transaction is reached from
			if spendingIn.Vout == index {
				continue
			}
			spentTx, e := db.GetTx(spendingIn.TxID)
			if e != nil {
				return nil, e
			}
			addr := spentTx.Vout[spendingIn.Vout].ScriptpubkeyAddress
			for _, inputAddr := range inputAddresses {
				if addr == inputAddr {
					c = []uint32{index}
					return
				}
			}
		}
	}

	err = errors.New("No output address matching forward heurisitic requirements found")
	return
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	return err == nil
}
