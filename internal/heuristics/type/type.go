// Package class heuristic
// This heuristic is the address type heuristic and it checks if the all the inputs 
// are of the same type and then try to locate only one output 
// that is of the same type. Again, we just need to check a simple condition.
package class

import (
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/txscript"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
)

// ChangeOutput returnes the index of the output which address type corresponds to input addresses type
func ChangeOutput(db storage.DB, tx *models.Tx) (uint32, error) {
	var inputTypes []txscript.ScriptClass
	var outputTypes []txscript.ScriptClass

	for _, in := range tx.Vin {
		if in.IsCoinbase {
			continue
		}
		spentTx, err := db.GetTx(in.TxID)
		if err != nil {
			return 0, err
		}
		script, _ := hex.DecodeString(spentTx.Vout[in.Vout].Scriptpubkey)
		class := txscript.GetScriptClass(script)
		inputTypes = append(inputTypes, class)
	}
	// check all inputs are of the same type
	for _, class := range inputTypes {
		if class.String() != inputTypes[0].String() {
			return 0, errors.New("There are different kind of addresses between inputs")
		}
	}
	for _, out := range tx.Vout {
		script, _ := hex.DecodeString(out.Scriptpubkey)
		class := txscript.GetScriptClass(script)
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
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	return err == nil
}
