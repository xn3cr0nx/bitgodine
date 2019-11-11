package forward

import (
	"errors"
	"fmt"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/dgraph"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
)

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(db *dgraph.Dgraph, tx *models.Tx) (uint32, error) {
	var inputAddresses []string

	logger.Debug("Forward Heuristic", fmt.Sprintf("transaction %s", tx.TxID), logger.Params{})

	for _, in := range tx.Vin {
		if in.IsCoinbase {
			continue
		}
		spentTx, err := db.GetTx(in.TxID)
		if err != nil {
			return 0, err
		}
		inputAddresses = append(inputAddresses, spentTx.Vout[in.Vout].ScriptpubkeyAddress)
	}

	for _, out := range tx.Vout {
		spendingTx, err := db.GetFollowingTx(tx.TxID, out.Index)
		if err != nil {
			// transaction not found => output not yet spent, but we can identify the change output anyway
			if err.Error() == "transaction not found" {
				continue
			}
			return 0, err
		}
		logger.Debug("Forward Heuristic", fmt.Sprintf("tx spending output vout %d: %s", out.Index, spendingTx.TxID), logger.Params{})
		for _, spendingIn := range spendingTx.Vin {
			logger.Debug("Forward Heuristic", fmt.Sprintf("input of spending tx %s", spendingIn.TxID), logger.Params{})
			// check if the input is the one the spending transaction is reached from
			if spendingIn.Vout == out.Index {
				continue
			}
			spentTx, err := db.GetTx(spendingIn.TxID)
			if err != nil {
				return 0, err
			}
			logger.Debug("Forward Heuristic", fmt.Sprintf("spent tx %s", spentTx.Vout), logger.Params{})
			addr := spentTx.Vout[spendingIn.Vout].ScriptpubkeyAddress
			for _, inputAddr := range inputAddresses {
				if addr == inputAddr {
					return out.Index, nil
				}
			}
		}
	}

	return 0, errors.New("No output address matching forward heurisitic requirements found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(db *dgraph.Dgraph, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	return err == nil
}
