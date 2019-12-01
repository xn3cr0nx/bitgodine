package backward

import (
	"errors"
	"fmt"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
)

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(dg storage.DB, tx *models.Tx) (uint32, error) {
	var outputAddresses,
		inputAddresses,
		inputTargets []string
	var spentTxs []models.Tx
	var outputTargets []uint32

	logger.Debug("Backward Heuristic", fmt.Sprintf("transaction %s", tx.TxID), logger.Params{})

	for _, out := range tx.Vout {
		outputAddresses = append(outputAddresses, out.ScriptpubkeyAddress)
	}
	for _, in := range tx.Vin {
		if in.IsCoinbase {
			continue
		}
		spentTx, err := dg.GetTx(in.TxID)
		if err != nil {
			return 0, err
		}
		addr := spentTx.Vout[in.Vout].ScriptpubkeyAddress
		inputAddresses = append(inputAddresses, addr)
		spentTxs = append(spentTxs, spentTx)
	}

	for _, spent := range spentTxs {
		logger.Debug("Backward Heuristic", fmt.Sprintf("spent transaction %s", spent.TxID), logger.Params{})

		for _, in := range spent.Vin {
			spentTx, err := dg.GetTx(in.TxID)
			if err != nil {
				return 0, err
			}
			if in.IsCoinbase {
				continue
			}
			addr := spentTx.Vout[in.Vout].ScriptpubkeyAddress
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
func Vulnerable(dg storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(dg, tx)
	return err == nil
}
