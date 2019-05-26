package backward

import (
	"errors"
	"fmt"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(tx *dgraph.Transaction) (uint32, error) {
	var outputAddresses,
		inputAddresses,
		inputTargets []string
	var spentTxs []dgraph.Transaction
	var outputTargets []uint32

	logger.Debug("Backward Heuristic", fmt.Sprintf("transaction %s", tx.Hash), logger.Params{})

	for _, out := range tx.Outputs {
		outputAddresses = append(outputAddresses, out.Address)
	}
	for _, in := range tx.Inputs {
		spentTx, err := dgraph.GetTx(in.Hash)
		if err != nil {
			return 0, err
		}
		addr := spentTx.Outputs[in.Vout].Address
		inputAddresses = append(inputAddresses, addr)
		spentTxs = append(spentTxs, spentTx)
	}

	for _, spent := range spentTxs {
		logger.Debug("Backward Heuristic", fmt.Sprintf("spent transaction %s", spent.Hash), logger.Params{})
		for _, in := range spent.Inputs {
			spentTx, err := dgraph.GetTx(in.Hash)
			if err != nil {
				return 0, err
			}
			addr := spentTx.Outputs[in.Vout].Address
			for _, inputAddr := range inputAddresses {
				if addr == inputAddr {
					inputTargets = append(inputTargets, addr)
				}
			}
			if len(inputTargets) > 0 {
				for target, outputAddr := range outputAddresses {
					if addr == outputAddr {
						outputTargets = append(outputTargets, uint32(target))
					}
				}

				for _, target := range outputTargets {
					for _, input := range inputTargets {
						if outputAddresses[int(target)] != input {
							return target, nil
						}
					}
				}
				inputTargets, outputTargets = []string{}, []uint32{}
			}
		}
	}

	return 0, errors.New("No output address matching backward heurisitic requirements found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(tx *dgraph.Transaction) bool {
	_, err := ChangeOutput(tx)
	if err == nil {
		return true
	}
	return false
}
