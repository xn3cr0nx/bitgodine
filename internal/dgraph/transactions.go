package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	// "github.com/btcsuite/btcd/chaincfg"
	// "github.com/btcsuite/btcd/txscript"
	// "github.com/btcsuite/btcd/wire"
	// "github.com/btcsuite/btcutil"
	// "github.com/dgraph-io/dgo/protos/api"
	// "github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Transaction represents the tx node structure in dgraph
type Transaction struct {
	UID      string   `json:"uid,omitempty"`
	Hash     string   `json:"hash,omitempty"`
	Locktime uint32   `json:"locktime,omitempty"`
	Version  int32    `json:"version,omitempty"`
	Inputs   []Input  `json:"inputs,omitempty"`
	Outputs  []Output `json:"outputs,omitempty"`
}

// Input represent input transaction, e.g. the link to a previous spent tx hash
type Input struct {
	UID             string      `json:"uid,omitempty"`
	Hash            string      `json:"hash,omitempty"`
	Vout            uint32      `json:"vout,omitempty"`
	SignatureScript string      `json:"signature_script,omitempty"`
	Witness         []TxWitness `json:"witness,omitempty"`
}

// TxWitness encodes witness slice into a string
type TxWitness string

// Output represent output transaction, e.g. the value that can be spent as input
type Output struct {
	UID      string `json:"uid,omitempty"`
	Value    int64  `json:"value,omitempty"`
	Vout     uint32 `json:"vout"`
	Address  string `json:"address,omitempty"`
	PkScript string `json:"pk_script,omitempty"`
}

// TxResp represent the resp from a dgraph query returning a transaction node
type TxResp struct {
	Txs []struct{ Transaction }
}

// OutputsResp represent the resp from a dgraph query returning an array of output nodes
type OutputsResp struct {
	Transactions []struct {
		Outputs []struct{ Output }
	}
}

// StoreCoinbase prepare coinbase output to be used as input for coinbase transactions
func StoreCoinbase() error {
	t := Transaction{
		Hash: strings.Repeat("0", 64),
		Outputs: []Output{
			Output{
				Value:    int64(5000000000),
				Address:  strings.Repeat("0", 64),
				Vout:     0,
				PkScript: "",
			},
			Output{
				Value:    int64(2500000000),
				Address:  strings.Repeat("0", 64),
				Vout:     0,
				PkScript: "",
			},
			Output{
				Value:    int64(1250000000),
				Address:  strings.Repeat("0", 64),
				Vout:     0,
				PkScript: "",
			},
		},
	}
	return Store(&t)
}

// GetTx returnes the node from the query queried
func GetTx(hash string) (Transaction, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		txs(func: eq(hash, "%s")) {
			uid
			hash
			locktime
			version
			inputs {
				uid
				hash
				vout
				signature_script
				witness
			}
			outputs {
				uid
				value
				vout
				address
				pk_script
			}
		}
	}`, hash))
	if err != nil {
		return Transaction{}, err
	}
	var r TxResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Transaction{}, err
	}
	if len(r.Txs) == 0 {
		return Transaction{}, errors.New("transaction not found")
	}
	for _, tx := range r.Txs {
		if len(tx.Transaction.Outputs) > 0 {
			return tx.Transaction, nil
		}
	}
	return Transaction{}, errors.New("transaction not found")
}

// GetTxUID returnes the uid of the queried tx by hash
func GetTxUID(hash *string) (string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		txs(func: allofterms(hash, %s)) {
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
	if len(r.Txs) == 0 {
		return "", errors.New("transaction not found")
	}

	return r.Txs[0].UID, nil
}

// GetTxOutputs returnes the outputs of the queried tx by hash
func GetTxOutputs(hash *string) ([]Output, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		transactions(func: allofterms(hash, %s)) {
			outputs {
				uid
        value
        vout
        address
        pk_script
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
	if len(r.Transactions[0].Outputs) == 0 {
		return []Output{}, errors.New("outputs not found")
	}
	var outputs []Output
	for _, o := range r.Transactions[0].Outputs {
		outputs = append(outputs, o.Output)
	}

	return outputs, nil
}

// GetSpentTxOutput returnes the output spent (the vout) of the corresponding tx
func GetSpentTxOutput(hash *string, vout *uint32) (Output, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		transactions(func: allofterms(hash, %s)) {
			outputs @filter(eq(vout, %d)) {
				uid
        value
        vout
        address
        pk_script
      }
		}
	}`, *hash, *vout))
	if err != nil {
		return Output{}, err
	}
	var r OutputsResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Output{}, err
	}
	if len(r.Transactions) == 0 {
		return Output{}, errors.New("output not found")
	}
	return r.Transactions[0].Outputs[0].Output, nil
}

// GetFollowingTx returns the transaction spending the output (vout) of
// the transaction passed as input to the function
func GetFollowingTx(hash *string, vout *uint32) (Transaction, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		txs(func: has(inputs)) @cascade {
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
	if len(r.Txs) == 0 {
		return Transaction{}, errors.New("transaction not found")
	}
	node := r.Txs[0].Transaction
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
	if len(r.Txs) == 0 {
		return nil, errors.New("No transaction stored")
	}
	var transactions []string
	for _, tx := range r.Txs {
		transactions = append(transactions, tx.Hash)
	}
	return transactions, nil
}

// GetTxBlockHeight returnes the height of the block based on its hash
func GetTxBlockHeight(hash string) (int32, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		block(func: has(prev_block)) @cascade {
			height
			transactions @filter(eq(hash, "%s"))
		}
	}`, hash))
	if err != nil {
		return 0, err
	}
	var r struct{ Block []struct{ Height int32 } }
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return 0, err
	}
	if len(r.Block) == 0 {
		return 0, errors.New("Block height not found")
	}
	return r.Block[0].Height, nil
}

// GetTransactionsHeightRange returnes the list of transaction contained between height boundaries passed as arguments
func GetTransactionsHeightRange(from, to *int32) ([]Transaction, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		txs(func: ge(height, %d), first: %d)  {
			transactions @filter(gt(count(outputs), 1)) {
        expand(_all_) {
          expand(_all_)
        }
      }
		}
	}`, *from, (*to)-(*from)+1))
	if err != nil {
		return nil, err
	}
	var r struct {
		Txs []struct {
			Transactions []struct {
				Transaction
			}
		}
	}
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, err
	}
	if len(r.Txs[0].Transactions) == 0 {
		return nil, errors.New("No transaction found in the block height range")
	}

	var txs []Transaction
	for _, block := range r.Txs {
		for _, tx := range block.Transactions {
			txs = append(txs, tx.Transaction)
		}
	}
	return txs, nil
}
