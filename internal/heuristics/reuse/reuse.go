package reuse

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
)

func contains(recipient []btcutil.Address, element btcutil.Address) bool {
	for _, v := range recipient {
		if v.String() == element.String() {
			return true
		}
	}
	return false
}

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(tx *txs.Tx) (uint32, error) {
	var inputAddresses []btcutil.Address

	for vout, in := range tx.MsgTx().TxIn {
		spentTx, err := tx.GetSpentTx(uint32(vout))
		if err != nil {
			if err.Error() == "Coinbase transaction" {
				continue
			}
			return 0, err
		}
		_, addr, _, err := txscript.ExtractPkScriptAddrs(spentTx.MsgTx().TxOut[in.PreviousOutPoint.Index].PkScript, &chaincfg.MainNetParams)
		if err != nil {
			return 0, err
		}
		inputAddresses = append(inputAddresses, addr[0])
	}

	// Here on the first matching output, that output is returned as change, but could be a reuse on more outputs?
	for vout, out := range tx.MsgTx().TxOut {
		_, addr, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
		if err != nil {
			return 0, err
		}
		if contains(inputAddresses, addr[0]) {
			return uint32(vout), nil
		}
	}

	return 0, errors.New("No reuse address found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(tx *txs.Tx) bool {
	_, err := ChangeOutput(tx)
	if err == nil {
		return true
	}
	return false
}
