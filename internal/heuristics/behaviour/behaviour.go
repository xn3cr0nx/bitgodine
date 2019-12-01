package behaviour

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_server/internal/address"
)

// ChangeOutput returnes the index of the output which appears for the first time in the chain based on client behaviour heuristic
func ChangeOutput(db storage.DB, tx *models.Tx) (uint32, error) {
	var outputAddresses []uint32
	blockHeight, err := db.GetTxBlockHeight(tx.TxID)
	if err != nil {
		return 0, err
	}
	for vout, out := range tx.Vout {
		// _, addr, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
		// if err != nil {
		// 	return 0, err
		// }
		var firstOccurence int32
		// if len(addr) > 0 {
		if out.ScriptpubkeyAddress != "" {
			addr, err := btcutil.DecodeAddress(out.ScriptpubkeyAddress, &chaincfg.MainNetParams)
			if err != nil {
				return 0, err
			}
			firstOccurence, err = address.FirstAppearence(db, &addr)
			if err != nil {
				return 0, err
			}
		} else {
			continue
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
func Vulnerable(db storage.DB, tx *models.Tx) bool {
	_, err := ChangeOutput(db, tx)
	return err == nil
}
