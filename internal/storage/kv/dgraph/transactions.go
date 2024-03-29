package dgraph

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// TxResp represent the resp from a dgraph query returning a transaction node
type TxResp struct {
	Txs []struct{ tx.Tx }
}

// OutputsResp represent the resp from a dgraph query returning an array of output nodes
type OutputsResp struct {
	Transactions []struct {
		Output []struct{ tx.Output }
	}
}

// Coinbase returns whether an input is refers to coinbase output
func Coinbase(in *tx.Input) bool {
	zeroHash, _ := chainhash.NewHash(make([]byte, 32))
	return in.TxID == zeroHash.String()
}

// StoreCoinbase prepare coinbase output to be used as input for coinbase transactions
func (d *Dgraph) StoreCoinbase() error {
	t := tx.Tx{
		TxID: strings.Repeat("0", 64),
		Vout: []tx.Output{
			{
				Scriptpubkey:        "",
				ScriptpubkeyAsm:     "",
				ScriptpubkeyType:    "",
				ScriptpubkeyAddress: strings.Repeat("0", 64),
				Value:               int64(5000000000),
				Index:               4294967295,
			},
			{
				Scriptpubkey:        "",
				ScriptpubkeyAsm:     "",
				ScriptpubkeyType:    "",
				ScriptpubkeyAddress: strings.Repeat("0", 64),
				Value:               int64(2500000000),
				Index:               4294967295,
			},
			{
				Scriptpubkey:        "",
				ScriptpubkeyAsm:     "",
				ScriptpubkeyType:    "",
				ScriptpubkeyAddress: strings.Repeat("0", 64),
				Value:               int64(1250000000),
				Index:               4294967295,
			},
		},
	}
	err := d.Store(&t)
	return err
}

// GetTx returns the node from the query queried
// TODO: orderasc on inputs, outputs, check whether they can have more than 1000 elments (1000 dgraph limit fetch)
func (d *Dgraph) GetTx(hash string) (transaction tx.Tx, err error) {
	if cached, ok := d.cache.Get(hash); ok {
		var r tx.Tx
		if err := json.Unmarshal(cached.([]byte), &r); err == nil {
			return r, nil
		}
	}

	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string) {
			txs(func: eq(txid, $s)) {
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
		}`, map[string]string{"$s": hash})
	if err != nil {
		return
	}
	var r TxResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}

	if len(r.Txs) == 0 {
		err = errorx.ErrTxNotFound
		return
	}
	for _, t := range r.Txs {
		if len(t.Tx.Vout) > 0 {

			bytes, err := json.Marshal(t.Tx)
			if err == nil {
				if !d.cache.Set(t.Tx.TxID, bytes, 1) {
					logger.Error("Cache", errorx.ErrCache, logger.Params{"hash": t.Tx.TxID})
				}
			}

			return t.Tx, nil
		}
	}

	err = errorx.ErrTxNotFound
	return
}

// GetTxUID returns the uid of the queried tx by hash
func (d *Dgraph) GetTxUID(hash string) (uid string, err error) {
	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string) {
			txs(func: allofterms(txid, $s)) @cascade {
				uid
				output {
					uid
				}
			}
		}`, map[string]string{"$s": hash})
	if err != nil {
		return
	}
	var r TxResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Txs) == 0 {
		err = errorx.ErrTxNotFound
		return
	}
	// uid = r.Txs[0].Tx.UID
	return
}

// GetTxOutputs returns the outputs of the queried tx by hash
// TODO: rememeber orderasc fetches no more than 1000 elements
func (d *Dgraph) GetTxOutputs(hash string) (outputs []tx.Output, err error) {
	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string) {
			transactions(func: allofterms(txid, $s)) {
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
		}`, map[string]string{"$s": hash})
	if err != nil {
		return
	}
	var r OutputsResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Transactions[0].Output) == 0 {
		err = fmt.Errorf("outputs %w", errorx.ErrNotFound)
		return
	}
	for _, o := range r.Transactions[0].Output {
		outputs = append(outputs, o.Output)
	}
	return
}

// GetSpentTxOutput returns the output spent (the vout) of the corresponding tx
func (d *Dgraph) GetSpentTxOutput(hash string, vout uint32) (output tx.Output, err error) {
	if cached, ok := d.cache.Get(fmt.Sprintf("%s_%d", hash, vout)); ok {
		var r tx.Output
		if err := json.Unmarshal(cached.([]byte), &r); err == nil {
			return r, nil
		}
	}
	if cached, ok := d.cache.Get(hash); ok {
		var r tx.Tx
		if err := json.Unmarshal(cached.([]byte), &r); err == nil {
			return r.Vout[vout], nil
		}
	}

	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string, $d: int) {
			transactions(func: allofterms(txid, $s)) {
				output @filter(eq(index, $d)) {
					uid
					scriptpubkey
					scriptpubkey_asm
					scriptpubkey_type
					scriptpubkey_address
					value
					index
  	    }
			}
		}`, map[string]string{"$s": hash, "$d": fmt.Sprintf("%d", vout)})
	if err != nil {
		return
	}
	var r OutputsResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}

	if len(r.Transactions) == 0 {
		err = fmt.Errorf("outputs %w", errorx.ErrNotFound)
		return
	}

	bytes, err := json.Marshal(r.Transactions[0].Output[0].Output)
	if err == nil {
		if !d.cache.Set(fmt.Sprintf("%s_%d", hash, vout), bytes, 1) {
			logger.Error("Cache", errorx.ErrCache, logger.Params{"vout": r.Transactions[0].Output[0].Output.Index})
		}
	}

	// why always the first
	return r.Transactions[0].Output[0].Output, nil
}

// GetFollowingTx returns the transaction spending the output (vout) of
// the transaction passed as input to the function
// TODO: rememeber orderasc fetches no more than 1000 elements
func (d *Dgraph) GetFollowingTx(hash string, vout uint32) (transaction tx.Tx, err error) {
	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string, $d: int) {
			txs(func: has(input)) @cascade {
				uid
				txid
				input @filter(eq(txid, $s) AND eq(vout, $d)){
					txid
					vout
				}
				output (orderasc: index) {
					value
					index
				}
			}
		}`, map[string]string{"$s": hash, "$d": fmt.Sprintf("%d", vout)})
	if err != nil {
		return
	}
	var r TxResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Txs) == 0 {
		err = errorx.ErrTxNotFound
		return
	}
	transaction = r.Txs[0].Tx
	return
}

// GetStoredTxs returns all the stored transactions hashes
func (d *Dgraph) GetStoredTxs() (transactions []string, err error) {
	resp, err := d.NewReadOnlyTxn().Query(context.Background(), `{
			txs(func: has(input)) {
				txid
			}
		}`)
	if err != nil {
		return
	}
	var r TxResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Txs) == 0 {
		err = errorx.ErrTxNotFound
	}
	for _, tx := range r.Txs {
		transactions = append(transactions, tx.Tx.TxID)
	}
	return
}

// GetTxBlockHeight returns the height of the block based on its hash
func (d *Dgraph) GetTxBlockHeight(hash string) (height int32, err error) {
	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string) {
			block(func: has(prev_block)) @cascade {
				height
				transactions @filter(eq(txid, "$s"))
			}
		}`, map[string]string{"$s": hash})
	if err != nil {
		return
	}
	var r struct{ Block []struct{ Height int32 } }
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Block) == 0 {
		err = errorx.ErrBlockNotFound
		return
	}
	height = r.Block[0].Height
	return
}

// GetTxBlock returns the block containing the transaction
func (d *Dgraph) GetTxBlock(hash string) (blk block.Block, err error) {
	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string) {
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
				transactions @filter(eq(txid, "$s"))
			}
		}`, map[string]string{"$s": hash})
	if err != nil {
		return
	}
	var r struct{ Block []struct{ block.Block } }
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Block) == 0 {
		err = errorx.ErrBlockNotFound
		return
	}
	blk = r.Block[0].Block
	return
}

// IsSpent returns true if exists a transaction that takes as input to the new tx
// the output corresponding to the index passed to the function
func (d *Dgraph) IsSpent(tx string, index uint32) bool {
	_, err := d.GetFollowingTx(tx, index)
	if err != nil {
		// just for sake of clarity, untill I'm going to refactor this piece to be more useful
		if errors.Is(err, errorx.ErrTxNotFound) {
			return false
		}
		return false
	}
	return true
}
