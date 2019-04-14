package backward

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(tx *txs.Tx) (uint32, error) {
	var outputAddresses,
		inputAddresses,
		inputTargets []btcutil.Address
	var spentTxs []txs.Tx
	var outputTargets []uint32

	logger.Debug("Backward Heuristic", fmt.Sprintf("transaction %s", tx.Hash().String()), logger.Params{})

	for _, out := range tx.MsgTx().TxOut {
		_, addr, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
		if err != nil {
			return 0, err
		}
		outputAddresses = append(outputAddresses, addr[0])
	}

	for vout, in := range tx.MsgTx().TxIn {
		spentTx, err := tx.GetSpentTx(uint32(vout))
		if err != nil {
			return 0, err
		}
		_, addr, _, err := txscript.ExtractPkScriptAddrs(spentTx.MsgTx().TxOut[int(in.PreviousOutPoint.Index)].PkScript, &chaincfg.MainNetParams)
		if err != nil {
			return 0, err
		}
		inputAddresses = append(inputAddresses, addr[0])
		spentTxs = append(spentTxs, spentTx)
	}

	for _, spent := range spentTxs {
		logger.Debug("Backward Heuristic", fmt.Sprintf("spent transaction %s", spent.Hash().String()), logger.Params{})
		for vout, in := range spent.MsgTx().TxIn {
			spentTx, err := spent.GetSpentTx(uint32(vout))
			if err != nil {
				return 0, err
			}
			_, addr, _, err := txscript.ExtractPkScriptAddrs(spentTx.MsgTx().TxOut[int(in.PreviousOutPoint.Index)].PkScript, &chaincfg.MainNetParams)
			if err != nil {
				return 0, err
			}
			for _, inputAddr := range inputAddresses {
				if addr[0].EncodeAddress() == inputAddr.EncodeAddress() {
					inputTargets = append(inputTargets, addr[0])
				}
			}
			if len(inputTargets) > 0 {
				for target, outputAddr := range outputAddresses {
					if addr[0].EncodeAddress() == outputAddr.EncodeAddress() {
						outputTargets = append(outputTargets, uint32(target))
					}
				}

				for _, target := range outputTargets {
					for _, input := range inputTargets {
						if outputAddresses[int(target)].EncodeAddress() != input.EncodeAddress() {
							return target, nil
						}
					}
				}
				inputTargets, outputTargets = []btcutil.Address{}, []uint32{}
			}
		}
	}

	return 0, errors.New("No output address matching backward heurisitic requirements found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(tx *txs.Tx) bool {
	_, err := ChangeOutput(tx)
	if err == nil {
		return true
	}
	return false
}
