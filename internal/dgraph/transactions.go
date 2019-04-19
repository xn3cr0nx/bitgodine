package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Transaction represents the tx node structure in dgraph
type Transaction struct {
	UID      string   `json:"uid,omitempty"`
	Hash     string   `json:"hash,omitempty"`
	Locktime uint32   `json:"locktime,omitempty"`
	Version  uint32   `json:"version,omitempty"`
	Inputs   []Input  `json:"inputs,omitempty"`
	Outputs  []Output `json:"outputs,omitempty"`
}

// Input represent input transaction, e.g. the link to a previous spent tx hash
type Input struct {
	UID  string `json:"uid,omitempty"`
	Hash string `json:"hash,omitempty"`
	Vout uint32 `json:"vout,omitempty"`
}

// Output represent output transaction, e.g. the value that can be spent as input
type Output struct {
	UID     string `json:"uid,omitempty"`
	Value   int64  `json:"value,omitempty"`
	Vout    uint32 `json:"vout"`
	Address string `json:"address,omitempty"`
}

// TxResp represent the resp from a dgraph query returning a transaction node
type TxResp struct {
	Tx []struct{ Transaction }
}

// OutputsResp represent the resp from a dgraph query returning an array of output nodes
type OutputsResp struct {
	Outputs []struct{ Output }
}

// PrepareTransactions parses the btcutil.TX array of structs and convert them in Transaction object compatible with dgraph schema
func PrepareTransactions(txs []*btcutil.Tx, height int32) ([]Transaction, error) {
	var transactions []Transaction

	for _, tx := range txs {
		inputs, err := prepareInputs(tx.MsgTx().TxIn, height)
		if err != nil {
			return nil, err
		}
		outputs, err := prepareOutputs(tx.MsgTx().TxOut)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, Transaction{
			Hash:     tx.Hash().String(),
			Locktime: tx.MsgTx().LockTime,
			Inputs:   inputs,
			Outputs:  outputs,
		})
	}

	return transactions, nil
}

func prepareInputs(inputs []*wire.TxIn, height int32) ([]Input, error) {
	var txIns []Input
	for _, in := range inputs {
		h := in.PreviousOutPoint.Hash.String()
		txOutputs, err := GetTxOutputs(&h)
		if err != nil {
			return nil, err
		}
		var spentOutput Output
		for _, out := range txOutputs {
			if in.PreviousOutPoint.Index == uint32(4294967295) {
				spendingCoinbase := blocks.CoinbaseValue(height)
				if out.Value == spendingCoinbase {
					spentOutput = out
					break
				}
			}
			if out.Vout == in.PreviousOutPoint.Index {
				spentOutput = out
				break
			}
		}
		if spentOutput.UID == "" {
			return nil, errors.New("something not working")
		}
		txIns = append(txIns, Input{UID: spentOutput.UID, Hash: h, Vout: in.PreviousOutPoint.Index})
	}
	return txIns, nil
}

func prepareOutputs(outputs []*wire.TxOut) ([]Output, error) {
	var txOuts []Output
	for k, out := range outputs {
		if out.PkScript == nil {
			// txOuts = append(txOuts, Output{UID: "_:output", Value: out.Value})
			txOuts = append(txOuts, Output{Value: out.Value})
		} else {
			// txOuts = append(txOuts, Output{UID: "_:output", Value: out.Value, Vout: uint32(k)})
			_, addr, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
			if err != nil {
				return nil, err
			}
			txOuts = append(txOuts, Output{Value: out.Value, Vout: uint32(k), Address: addr[0].EncodeAddress()})
		}
	}

	return txOuts, nil
}

// StoreTx stores bitcoin transaction in the graph
func StoreTx(tx *btcutil.Tx, height int32) error {
	// check if tx is already stored
	hash := tx.Hash().String()
	if _, err := GetTxUID(&hash); err == nil {
		logger.Debug("Dgraph", "already stored transaction", logger.Params{"hash": hash})
		return nil
	}

	txIns, err := prepareInputs(tx.MsgTx().TxIn, height)
	if err != nil {
		return err
	}
	txOuts, err := prepareOutputs(tx.MsgTx().TxOut)
	if err != nil {
		return err
	}

	t := Transaction{
		Hash:     hash,
		Locktime: tx.MsgTx().LockTime,
		Inputs:   txIns,
		Outputs:  txOuts,
	}
	out, err := json.Marshal(t)
	if err != nil {
		return err
	}
	resp, err := instance.NewTxn().Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	if err != nil {
		return err
	}
	logger.Debug("Dgraph", resp.String(), logger.Params{})
	return nil
}

// GetTx returnes the node from the query queried
func GetTx(field string, param *string) (Transaction, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		transaction(func: allofterms(%s, %s)) {
			expand(_all_) {
				expand(_all_)
			}
		}
	}`, field, *param))
	if err != nil {
		return Transaction{}, err
	}
	var r TxResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Transaction{}, err
	}
	if len(r.Tx) == 0 {
		return Transaction{}, errors.New("transaction not found")
	}

	return r.Tx[0].Transaction, nil
}

// GetTxUID returnes the uid of the queried tx by hash
func GetTxUID(hash *string) (string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		tx(func: allofterms(hash, %s)) {
			uid
		}
	}`, *hash))
	if err != nil {
		return "", err
	}
	var r TxResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", err
	}
	if len(r.Tx) == 0 {
		return "", errors.New("transaction not found")
	}

	return r.Tx[0].UID, nil
}

// GetTxOutputs returnes the outputs of the queried tx by hash
func GetTxOutputs(hash *string) ([]Output, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		outputs(func: allofterms(hash, %s)) {
			outputs {
        expand(_all_)
      }
		}
	}`, *hash))
	if err != nil {
		return []Output{}, err
	}
	var r OutputsResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return []Output{}, err
	}
	if len(r.Outputs) == 0 {
		return []Output{}, errors.New("outputs not found")
	}
	var outputs []Output
	for _, o := range r.Outputs {
		outputs = append(outputs, o.Output)
	}

	return outputs, nil
}

// GetFollowingTx returns the transaction spending the output (vout) of
// the transaction passed as input to the function
func GetFollowingTx(hash *string, vout *uint32) (Transaction, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		tx(func: has(inputs)) @cascade {
			uid
			block
			hash
			inputs @filter(eq(hash, %s) AND eq(vout, %d)){
				hash
				vout
			}
			outputs {
				value
				vout
			}
		}
	}`, *hash, *vout))
	if err != nil {
		return Transaction{}, err
	}
	var r TxResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Transaction{}, err
	}
	if len(r.Tx) == 0 {
		return Transaction{}, errors.New("transaction not found")
	}
	node := r.Tx[0].Transaction
	return node, nil
}

// GetStoredTxs returnes all the stored transactions hashes
func GetStoredTxs() ([]string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), `{
			tx(func: has(inputs)) {
				hash
			}
		}`)
	if err != nil {
		return nil, err
	}
	var r TxResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, err
	}
	if len(r.Tx) == 0 {
		return nil, errors.New("No transaction stored")
	}
	var transactions []string
	for _, tx := range r.Tx {
		transactions = append(transactions, tx.Hash)
	}
	return transactions, nil
}
