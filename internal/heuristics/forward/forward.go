package forward

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
)

// ChangeOutput returnes the index of the output which appears both in inputs and in outputs based on address reuse heuristic
func ChangeOutput(tx *txs.Tx) (uint32, error) {
	var inputAddresses []btcutil.Address
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
	}

	for vout := range tx.MsgTx().TxOut {
		spendingTx, err := tx.GetSpendingTx(uint32(vout))
		if err != nil {
			// transaction not found => output not yet spent, but we can identify the change output anyway
			if err.Error() == "transaction not found" {
				continue
			}
			return 0, err
		}
		for in, spendingIn := range spendingTx.MsgTx().TxIn {
			spentTx, err := tx.GetSpentTx(uint32(in))
			if err != nil {
				return 0, err
			}
			_, addr, _, err := txscript.ExtractPkScriptAddrs(spentTx.MsgTx().TxOut[int(spendingIn.PreviousOutPoint.Index)].PkScript, &chaincfg.MainNetParams)
			if err != nil {
				return 0, err
			}
			for _, inputAddr := range inputAddresses {
				if addr[0].EncodeAddress() == inputAddr.EncodeAddress() {
					return uint32(vout), nil
				}
			}
		}
	}

	return 0, errors.New("No output address matching forward heurisitic requirements found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(tx *txs.Tx) bool {
	_, err := ChangeOutput(tx)
	if err == nil {
		return true
	}
	return false
}
