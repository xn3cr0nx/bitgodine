package dgraph

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// BlockResp represent the resp from a dgraph query returning a transaction node
type BlockResp struct {
	Blk []struct{ models.Block }
}

// StoreBlock returns the hash of the block retrieving it based on its height
func (d *Dgraph) StoreBlock(v interface{}) (err error) {
	b := v.(*models.Block)
	err = d.Store(b)
	return
}

// GetBlockFromHash returns the hash of the block retrieving it based on its height
func (d *Dgraph) GetBlockFromHash(hash string) (block models.Block, err error) {
	if cached, ok := d.cache.Get(hash); ok {
		block = cached.(models.Block)
		return
	}

	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string) {
			blk(func: eq(id, $s)) {
				uid
				id
				height
				version
				previousblockhash
				timestamp
				merkle_root
				bits
				nonce
				transactions {
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
			}
		}`, map[string]string{"$s": hash})
	if err != nil {
		return
	}
	var r BlockResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Blk) == 0 {
		err = errors.New("Block not found")
		return
	}
	if err == nil {
		if !d.cache.Set(r.Blk[0].Block.ID, r.Blk[0].Block, 1) {
			logger.Error("Cache", errors.New("error caching"), logger.Params{"hash": r.Blk[0].Block.ID})
		}
	}
	block, err = r.Blk[0].Block, nil
	return
}

// GetBlockFromHeight returns the hash of the block retrieving it based on its height
func (d *Dgraph) GetBlockFromHeight(height int32) (block models.Block, err error) {
	if cached, ok := d.cache.Get(height); ok {
		block = cached.(models.Block)
		return
	}

	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($d: int) {
			blk(func: eq(height, $d), first: 1) {
				uid
				id
				height
				version
				previousblockhash
				timestamp
				merkle_root
				bits
				nonce
				transactions {
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
			}
		}`, map[string]string{"$d": fmt.Sprintf("%d", height)})
	if err != nil {
		return
	}
	var r BlockResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Blk) == 0 {
		err = errors.New("Block not found")
		return
	}
	if err == nil {
		if !d.cache.Set(r.Blk[0].Block.Height, r.Blk[0].Block, 1) {
			logger.Error("Cache", errors.New("error caching"), logger.Params{"hash": r.Blk[0].Block.ID})
		}
	}
	block, err = r.Blk[0].Block, nil
	return
}

// GetBlockFromHeightRange returns the hash of the block retrieving it based on its height
func (d *Dgraph) GetBlockFromHeightRange(height int32, first int) (blocks []models.Block, err error) {
	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($d: int, $f: int) {
			blk(func: ge(height, $d), first: $f) {
				uid
				id
				height
				version
				previousblockhash
				timestamp
				merkle_root
				bits
				nonce
				transactions {
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
			}
		}`, map[string]string{"$d": fmt.Sprintf("%d", height), "$f": fmt.Sprintf("%d", first)})
	if err != nil {
		return
	}
	var r BlockResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Blk) == 0 {
		err = errors.New("Block not found")
		return
	}
	for _, b := range r.Blk {
		blocks = append(blocks, b.Block)
	}
	return
}

// GetLastBlockHeight returns the height of the last block synced by Bitgodine
func (d *Dgraph) GetLastBlockHeight() (height int32, err error) {
	resp, err := d.NewReadOnlyTxn().Query(context.Background(), `{
		var(func: has(previousblockhash)) {
			blocks_height as height
		}
		h() {
			height: max(val(blocks_height))
		}
	}`)
	if err != nil {
		return
	}

	type HeightResponse struct {
		H []struct {
			Height int32 `json:"height,omitempty"`
		} `json:"h,omitempty"`
	}

	var r HeightResponse
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		logger.Error("Dgraph Blocks", err, logger.Params{})
		if strings.Contains(err.Error(), "0.000000") {
			return 0, nil
		}
		return
	}
	if len(r.H) != 1 {
		err = errors.New("Something went wrong retrieving max height")
		return
	}
	height = r.H[0].Height
	return
}

// LastBlock returns the last block synced by Bitgodine
func (d *Dgraph) LastBlock() (block models.Block, err error) {
	resp, err := d.NewReadOnlyTxn().Query(context.Background(), `{
		var(func: has(previousblockhash)) {
			blocks_height as height
		}
		var() {
			h as max(val(blocks_height))
		}
    b(func: eq(height, val(h))) @filter(has(previousblockhash)) {
			uid
      id
			height
			version
			previousblockhash
			merkle_root
      timestamp
      bits
      nonce
    }
	}`)
	if err != nil {
		return
	}

	var r struct{ B []struct{ models.Block } }
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.B) != 1 {
		err = errors.New("Something went wrong retrieving last block")
		return
	}
	block = r.B[0].Block
	return
}

// GetStoredBlocks returns an array containing all blocks stored on dgraph
func (d *Dgraph) GetStoredBlocks() (blocks []models.Block, err error) {
	resp, err := d.NewReadOnlyTxn().Query(context.Background(), `{
		blocks(func: has(previousblockhash)) {
			height
			id
		}
	}`)
	if err != nil {
		return
	}

	var r struct{ Blocks []struct{ models.Block } }
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	for _, b := range r.Blocks {
		blocks = append(blocks, b.Block)
	}

	return
}

// // RemoveBlock removes the block specified by its height
// func (d *Dgraph) RemoveBlock(block *models.Block) error {
// 	return d.Delete(block.UID)
// }

// RemoveLastBlock deletes the last block stored in the db
func (d *Dgraph) RemoveLastBlock() (err error) {
	// var blocks []models.Block
	// var height int32
	// block, err := d.LastBlock()
	// if err != nil {
	// 	if err.Error() == "Something went wrong retrieving last block" {
	// 		height, err = d.GetLastBlockHeight()
	// 		if err != nil {
	// 			return
	// 		}
	// 		uids, err := d.getBlockUIDFromHeight(height)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		for _, uid := range uids {
	// 		blocks = append(blocks, models.Block{UID: uid})
	// 		}
	// 	} else {
	// 		return
	// 	}
	// }

	// if block.ID != "" {
	// 	if err = d.RemoveBlock(&block); err != nil {
	// 		return
	// 	}
	// 	logger.Info("Block rm", fmt.Sprintf("Block %d correctly removed", block.Height), logger.Params{})
	// } else {
	// 	for _, b := range blocks {
	// 		if err = d.RemoveBlock(&b); err != nil {
	// 			return
	// 		}
	// 	}
	// 	logger.Info("Block rm", fmt.Sprintf("Block %d correctly removed", height), logger.Params{})
	// }
	return
}

// getBlockUIDFromHeight returns the dgraph uid of the block stored at the passed height
func (d *Dgraph) getBlockUIDFromHeight(height int32) (uids []string, err error) {
	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `{
		block(func: eq(height, $d)) {
			uid
		}
	}`, map[string]string{"$d": fmt.Sprintf("%d", height)})
	if err != nil {
		return
	}
	var r struct {
		Block []struct {
			UID string `json:"uid,omitempty"`
		}
	}
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Block) == 0 {
		err = errors.New("Block not found")
		return
	}
	for _, b := range r.Block {
		uids = append(uids, b.UID)
	}
	return
}

// GetBlockTxOutputsFromHash retrieves the list of outputs uid of all block's transactions
func (d *Dgraph) GetBlockTxOutputsFromHash(hash string) (uids map[string][]string, err error) {
	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($s: string) {
			blk(func: eq(id, $s)) {
				transactions {
					txid
					output {
						uid
						index
					}
				}
			}
		}`, map[string]string{"$s": hash})
	if err != nil {
		if err != nil && strings.Contains(err.Error(), "transport is closing") {
			time.Sleep(2 * time.Second)
			uids, err = d.GetBlockTxOutputsFromHash(hash)
			return
		}
		return
	}
	var r BlockResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Blk) == 0 {
		err = errors.New("Block not found")
		return
	}

	// uids = make(map[string][]string)
	// for _, tx := range r.Blk[0].Block.Transactions {
	// 	uids[tx.TxID] = make([]string, len(tx.Vout))
	// 	for _, output := range tx.Vout {
	// 		uids[tx.TxID][output.Index] = output.UID
	// 	}
	// }

	return
}

// GetBlockTxOutputsFromRange retrieves the list of outputs uid of all block's transactions in the block range
func (d *Dgraph) GetBlockTxOutputsFromRange(height int32, first int) (uids map[string]map[string][]string, err error) {
	resp, err := d.NewReadOnlyTxn().QueryWithVars(context.Background(), `
		query params($d: int, $f: int) {
			blk(func: ge(height, $d), first: $f) {
				id
				height
				transactions {
					txid
					output {
						uid
						index
					}
				}
			}
		}`, map[string]string{"$d": fmt.Sprintf("%d", height), "$f": fmt.Sprintf("%d", first)})
	if err != nil {
		return
	}
	var r BlockResp
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return
	}
	if len(r.Blk) == 0 {
		err = errors.New("Block not found")
		return
	}

	// uids = make(map[string]map[string][]string)
	// for _, b := range r.Blk {
	// 	uids[b.ID] = make(map[string][]string)
	// 	for _, tx := range b.Block.Transactions {
	// 		uids[b.ID][tx.TxID] = make([]string, len(tx.Vout))
	// 		for _, output := range tx.Vout {
	// 			uids[b.ID][tx.TxID][output.Index] = output.UID
	// 		}
	// 	}
	// }

	return
}
