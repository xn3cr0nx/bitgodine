package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/allegro/bigcache"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/xn3cr0nx/bitgodine_code/internal/cache"
	"github.com/xn3cr0nx/bitgodine_code/internal/models"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// TxResp represent the resp from a dgraph query returning a transaction node
type TxResp struct {
	Txs []struct{ Tx models.Tx }
}

// OutputsResp represent the resp from a dgraph query returning an array of output nodes
type OutputsResp struct {
	Transactions []struct {
		Output []struct{ Output models.Output }
	}
}

// Coinbase returnes whether an input is refers to coinbase output
func Coinbase(in *models.Input) bool {
	zeroHash, _ := chainhash.NewHash(make([]byte, 32))
	return in.TxID == zeroHash.String()
}

// StoreCoinbase prepare coinbase output to be used as input for coinbase transactions
func StoreCoinbase() error {
	t := models.Tx{
		TxID: strings.Repeat("0", 64),
		Vout: []models.Output{
			models.Output{
				Scriptpubkey:        "",
				ScriptpubkeyAsm:     "",
				ScriptpubkeyType:    "",
				ScriptpubkeyAddress: strings.Repeat("0", 64),
				Value:               int64(5000000000),
				Index:               0,
			},
			models.Output{
				Scriptpubkey:        "",
				ScriptpubkeyAsm:     "",
				ScriptpubkeyType:    "",
				ScriptpubkeyAddress: strings.Repeat("0", 64),
				Value:               int64(2500000000),
				Index:               0,
			},
			models.Output{
				Scriptpubkey:        "",
				ScriptpubkeyAsm:     "",
				ScriptpubkeyType:    "",
				ScriptpubkeyAddress: strings.Repeat("0", 64),
				Value:               int64(1250000000),
				Index:               0,
			},
		},
	}
	return Store(&t)
}

// GetTx returnes the node from the query queried
// TODO: orderasc on inputs, outputs, check whether they can have more than 1000 elments (1000 dgraph limit fetch)
func GetTx(hash string) (models.Tx, error) {
	c, err := cache.Instance(bigcache.Config{})
	if err != nil {
		return models.Tx{}, err
	}
	cached, err := c.Get(hash)
	if len(cached) != 0 {
		var r models.Tx
		if err := json.Unmarshal(cached, &r); err == nil {
			return r, nil
		}
	}

	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		txs(func: eq(txid, "%s")) {
			uid
			txid
			version
			locktime
			size
			weight
			fee
			input (orderasc: vout) {
				uid
				txid
				vout
				is_coinbase
				scriptsig
				scriptsig_asm
				inner_redeemscript_asm
				inner_witnessscript_asm
				sequence
				witness
				prevout
			}
			output (orderasc: index) {
				uid
				scriptpubkey
				scriptpubkey_asm
				scriptpubkey_type
				scriptpubkey_address
				value
				index
			}
			status {
				uid
				confirmed
				block_height
				block_hash
				block_time
			}
		}
	}`, hash))
	if err != nil {
		return models.Tx{}, err
	}
	var r TxResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return models.Tx{}, err
	}
	if len(r.Txs) == 0 {
		return models.Tx{}, errors.New("transaction not found")
	}
	for _, tx := range r.Txs {
		if len(tx.Tx.Vout) > 0 {

			bytes, err := json.Marshal(tx.Tx)
			if err == nil {
				if err := c.Set(tx.Tx.TxID, bytes); err != nil {
					logger.Error("Cache", err, logger.Params{})
				}
			}

			return tx.Tx, nil
		}
	}

	return models.Tx{}, errors.New("transaction not found")
}

// GetTxUID returnes the uid of the queried tx by hash
func GetTxUID(hash *string) (string, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
			txs(func: allofterms(txid, %s)) @cascade {
				uid
				output {
					uid
				}
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
	return r.Txs[0].Tx.UID, nil
}

// GetTxOutputs returnes the outputs of the queried tx by hash
// TODO: rememeber orderasc fetches no more than 1000 elements
func GetTxOutputs(hash *string) ([]models.Output, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		transactions(func: allofterms(txid, %s)) {
			output (orderasc: index) {
				uid
				scriptpubkey
				scriptpubkey_asm
				scriptpubkey_type
				scriptpubkey_address
				value
				index
      }
		}
	}`, *hash))
	if err != nil {
		return []models.Output{}, err
	}
	var r OutputsResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return []models.Output{}, err
	}
	if len(r.Transactions[0].Output) == 0 {
		return []models.Output{}, errors.New("outputs not found")
	}
	var outputs []models.Output
	for _, o := range r.Transactions[0].Output {
		outputs = append(outputs, o.Output)
	}

	return outputs, nil
}

// GetSpentTxOutput returnes the output spent (the vout) of the corresponding tx
func GetSpentTxOutput(hash *string, vout *uint32) (models.Output, error) {
	c, err := cache.Instance(bigcache.Config{})
	if err != nil {
		return models.Output{}, err
	}
	cached, err := c.Get(fmt.Sprintf("%s_%d", *hash, *vout))
	if len(cached) != 0 {
		var r models.Output
		if err := json.Unmarshal(cached, &r); err == nil {
			return r, nil
		}
	}
	cached, err = c.Get(*hash)
	if len(cached) != 0 {
		var r models.Tx
		if err := json.Unmarshal(cached, &r); err == nil {
			return r.Vout[*vout], nil
		}
	}

	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		transactions(func: allofterms(txid, %s)) {
			output @filter(eq(index, %d)) {
				uid
				scriptpubkey
				scriptpubkey_asm
				scriptpubkey_type
				scriptpubkey_address
				value
				index
      }
		}
	}`, *hash, *vout))
	if err != nil {
		return models.Output{}, err
	}
	var r OutputsResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return models.Output{}, err
	}
	if len(r.Transactions) == 0 {
		return models.Output{}, errors.New("output not found")
	}

	bytes, err := json.Marshal(r.Transactions[0].Output[0].Output)
	if err == nil {
		if err := c.Set(fmt.Sprintf("%s_%d", *hash, *vout), bytes); err != nil {
			logger.Error("Cache", err, logger.Params{})
		}
	}

	return r.Transactions[0].Output[0].Output, nil
}

// GetFollowingTx returns the transaction spending the output (vout) of
// the transaction passed as input to the function
// TODO: rememeber orderasc fetches no more than 1000 elements
func GetFollowingTx(hash *string, vout *uint32) (models.Tx, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		txs(func: has(input)) @cascade {
			uid
			txid
			input @filter(eq(txid, %s) AND eq(vout, %d)){
				txid
				vout
			}
			output (orderasc: index) {
				value
				index
			}
		}
	}`, *hash, *vout))
	if err != nil {
		return models.Tx{}, err
	}
	var r TxResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return models.Tx{}, err
	}
	if len(r.Txs) == 0 {
		return models.Tx{}, errors.New("transaction not found")
	}
	node := r.Txs[0].Tx
	return node, nil
}

// GetStoredTxs returnes all the stored transactions hashes
func GetStoredTxs() ([]string, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), `{
			tx(func: has(input)) {
				txid
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
		transactions = append(transactions, tx.Tx.TxID)
	}
	return transactions, nil
}

// GetTxBlockHeight returnes the height of the block based on its hash
func GetTxBlockHeight(hash string) (int32, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		block(func: has(prev_block)) @cascade {
			height
			transactions @filter(eq(txid, "%s"))
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

// GetTxBlock returnes the block containing the transaction
func GetTxBlock(hash string) (Block, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		block(func: has(prev_block)) @cascade {
			uid
			txid
			height
			prev_block
			time
			version
			merkle_root
			bits
			nonce
			transactions @filter(eq(txid, "%s"))
		}
	}`, hash))
	if err != nil {
		return Block{}, err
	}
	var r struct{ Block []struct{ Block } }
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Block{}, err
	}
	if len(r.Block) == 0 {
		return Block{}, errors.New("Block not found")
	}
	return r.Block[0].Block, nil
}

// GetTransactionsHeightRange returnes the list of transaction contained between height boundaries passed as arguments
func GetTransactionsHeightRange(from, to *int32) ([]models.Tx, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		txs(func: ge(height, %d), first: %d)  {
			transactions @filter(gt(count(output), 1)) {
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
				Tx models.Tx
			}
		}
	}
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, err
	}
	if len(r.Txs) < 1 {
		return nil, errors.New("No transaction found in the block height range")
	}
	if len(r.Txs[0].Transactions) == 0 {
		return nil, errors.New("No transaction found in the block height range")
	}

	var txs []models.Tx
	for _, block := range r.Txs {
		for _, tx := range block.Transactions {
			txs = append(txs, tx.Tx)
		}
	}
	return txs, nil
}

// IsSpent returnes true if exists a transaction that takes as input to the new tx
// the output corresponding to the index passed to the function
func IsSpent(tx string, index uint32) bool {
	_, err := GetFollowingTx(&tx, &index)
	if err != nil {
		// just for sake of clarity, untill I'm going to refactor this piece to be more useful
		if err.Error() == "transaction not found" {
			return false
		}
		return false
	}
	return true
}
