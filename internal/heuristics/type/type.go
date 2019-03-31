package class

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
)

// ChangeOutput returnes the index of the output which address type corresponds to input addresses type
func ChangeOutput(tx *txs.Tx) (uint32, error) {
	var inputTypes []txscript.ScriptClass
	var outputTypes []txscript.ScriptClass

	for vout, in := range tx.MsgTx().TxIn {
		spentTx, err := tx.GetSpentTx(uint32(vout))
		if err != nil {
			if err.Error() == "Coinbase transaction" {
				continue
			}
			return 0, err
		}
		class, _, _, err := txscript.ExtractPkScriptAddrs(spentTx.MsgTx().TxOut[in.PreviousOutPoint.Index].PkScript, &chaincfg.MainNetParams)
		if err != nil {
			if err.Error() == "Coinbase transaction" {
				continue
			}
			return 0, err
		}
		fmt.Println("input class", class)
		inputTypes = append(inputTypes, class)
	}
	// check all inputs are of the same type
	for _, class := range inputTypes {
		if class.String() != inputTypes[0].String() {
			return 0, errors.New("There are different kind of addresses between inputs")
		}
	}
	for _, out := range tx.MsgTx().TxOut {
		class, _, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
		if err != nil {
			return 0, err
		}
		fmt.Println("output class", class)
		outputTypes = append(outputTypes, class)
	}
	// check there are not two or more outputs of the same type
	for k, class := range outputTypes {
		if k > 0 && class.String() == outputTypes[0].String() {
			return 0, errors.New("Two or more output of the same type, cannot determine change output")
		}
	}

	for _, input := range inputTypes {
		for vout, output := range outputTypes {
			if input.String() == output.String() {
				return uint32(vout), nil
			}
		}
	}
	return 0, errors.New("No address type matching input addresses type found")
}

// Vulnerable returnes true if the transaction has a privacy vulnerability due to optimal change heuristic
func Vulnerable(tx *txs.Tx) bool {
	_, err := ChangeOutput(tx)
	if err == nil {
		return true
	}
	return false
}
