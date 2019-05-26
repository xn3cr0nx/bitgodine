package forward

import (
	"errors"
	"fmt"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(tx *dgraph.Transaction) (uint32, error) {
	var inputAddresses []string

	logger.Debug("Forward Heuristic", fmt.Sprintf("transaction %s", tx.Hash), logger.Params{})

	for _, in := range tx.Inputs {
		spentTx, err := dgraph.GetTx(in.Hash)
		if err != nil {
			return 0, err
		}
		inputAddresses = append(inputAddresses, spentTx.Outputs[in.Vout].Address)
	}

	for _, out := range tx.Outputs {
		spendingTx, err := dgraph.GetFollowingTx(&tx.Hash, &out.Vout)
		if err != nil {
			// transaction not found => output not yet spent, but we can identify the change output anyway
			if err.Error() == "transaction not found" {
				continue
			}
			return 0, err
		}
		logger.Debug("Forward Heuristic", fmt.Sprintf("tx spending output vout %d: %s", out.Vout, spendingTx.Hash), logger.Params{})
		for _, spendingIn := range spendingTx.Inputs {
			logger.Debug("Forward Heuristic", fmt.Sprintf("input of spending tx %s", spendingIn.Hash), logger.Params{})
			// check if the input is the one the spending transaction is reached from
			if spendingIn.Vout == out.Vout {
				continue
			}
			spentTx, err := dgraph.GetTx(spendingIn.Hash)
			if err != nil {
				return 0, err
			}
			logger.Debug("Forward Heuristic", fmt.Sprintf("spent tx %s", spentTx.Hash), logger.Params{})
			addr := spentTx.Outputs[spendingIn.Vout].Address
			for _, inputAddr := range inputAddresses {
				if addr == inputAddr {
					return out.Vout, nil
				}
			}
		}
	}

	return 0, errors.New("No output address matching forward heurisitic requirements found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(tx *dgraph.Transaction) bool {
	_, err := ChangeOutput(tx)
	if err == nil {
		return true
	}
	return false
}
