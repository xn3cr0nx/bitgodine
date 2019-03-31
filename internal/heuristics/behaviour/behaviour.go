package behaviour

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/xn3cr0nx/bitgodine_code/internal/addresses"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
)

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(tx *txs.Tx) (uint32, error) {
	var outputAddresses []uint32
	blockHeight, err := tx.BlockHeight()
	if err != nil {
		return 0, err
	}
	for vout, out := range tx.MsgTx().TxOut {
		_, addr, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
		if err != nil {
			return 0, err
		}
		firstOccurence, err := addresses.FirstAppearence(addr[0].EncodeAddress())
		if err != nil {
			return 0, err
		}
		// check if occurence is the first, e.g. the transaction block height is the firstOccurence
		if firstOccurence == blockHeight {
			outputAddresses = append(outputAddresses, uint32(vout))
		}
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
func Vulnerable(tx *txs.Tx) bool {
	_, err := ChangeOutput(tx)
	if err == nil {
		return true
	}
	return false
}
