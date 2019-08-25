package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/allegro/bigcache"
	"github.com/xn3cr0nx/bitgodine_code/internal/cache"
	"github.com/xn3cr0nx/bitgodine_code/internal/models"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Block represent the dgraph node containing essential block info
type Block struct {
	UID          string      `json:"uid,omitempty"`
	ID           string      `json:"id,omitempty"`
	Height       int32       `json:"height,omitempty"`
	Version      int32       `json:"version,omitempty"`
	Timestamp    time.Time   `json:"timestamp,omitempty"`
	Bits         uint32      `json:"bits,omitempty"`
	Nonce        uint32      `json:"nonce,omitempty"`
	MerkleRoot   string      `json:"merkle_root,omitempty"`
	Transactions []models.Tx `json:"transactions,omitempty"`
	TxCount      int         `json:"tx_count,omitempty"`
	Size         int         `json:"size,omitempty"`
	Weight       int         `json:"weight,omitempty"`
	PrevBlock    string      `json:"prev_block,omitempty"`
}

// BlockResp represent the resp from a dgraph query returning a transaction node
type BlockResp struct {
	Blk []struct{ Block }
}

// GetBlockFromHash returnes the hash of the block retrieving it based on its height
func GetBlockFromHash(hash string) (Block, error) {
	c, err := cache.Instance(bigcache.Config{})
	if err != nil {
		return Block{}, err
	}
	cached, err := c.Get(hash)
	if len(cached) != 0 {
		var r Block
		if err := json.Unmarshal(cached, &r); err == nil {
			return r, nil
		}
	}

	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		blk(func: eq(id, %s)) {
			uid
			id
			height
			version
			prev_block
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
	}`, hash))
	if err != nil {
		return Block{}, err
	}
	var r BlockResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Block{}, err
	}
	if len(r.Blk) == 0 {
		return Block{}, errors.New("Block not found")
	}
	bytes, err := json.Marshal(r.Blk[0].Block)
	if err == nil {
		if err := c.Set(r.Blk[0].Block.ID, bytes); err != nil {
			logger.Error("Cache", err, logger.Params{})
		}
	}
	return r.Blk[0].Block, nil
}

// GetBlockFromHeight returnes the hash of the block retrieving it based on its height
func GetBlockFromHeight(height int32) (Block, error) {
	c, err := cache.Instance(bigcache.Config{})
	if err != nil {
		return Block{}, err
	}
	cached, err := c.Get(strconv.Itoa(int(height)))
	if len(cached) != 0 {
		var r Block
		if err := json.Unmarshal(cached, &r); err == nil {
			return r, nil
		}
	}

	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		blk(func: eq(height, %d), first: 1) {
			uid
			id
			height
			version
			prev_block
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
	}`, height))
	if err != nil {
		return Block{}, err
	}
	var r BlockResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Block{}, err
	}
	if len(r.Blk) == 0 {
		return Block{}, errors.New("Block not found")
	}
	bytes, err := json.Marshal(r.Blk[0].Block)
	if err == nil {
		if err := c.Set(strconv.Itoa(int(r.Blk[0].Block.Height)), bytes); err != nil {
			logger.Error("Cache", err, logger.Params{})
		}
	}
	return r.Blk[0].Block, nil
}

// GetBlockFromHeightRange returnes the hash of the block retrieving it based on its height
func GetBlockFromHeightRange(height int32, first int) ([]Block, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		blk(func: eq(height, %d), first: %d) {
			uid
			id
			height
			version
			prev_block
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
	}`, height, first))
	if err != nil {
		return nil, err
	}
	var r BlockResp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, err
	}
	if len(r.Blk) == 0 {
		return nil, errors.New("Block not found")
	}
	var blocks []Block
	for _, b := range r.Blk {
		blocks = append(blocks, b.Block)
	}
	return blocks, nil
}

// LastBlockHeight returnes the height of the last block synced by Bitgodine
func LastBlockHeight() (int32, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), `{
		var(func: has(prev_block)) {
			blocks_height as height
		}
		h() {
			height: max(val(blocks_height))
		}
	}`)
	if err != nil {
		return 0, err
	}

	type HeightResponse struct {
		H []struct {
			Height int32 `json:"height,omitempty"`
		} `json:"h,omitempty"`
	}

	var r HeightResponse
	// var r interface{}
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		logger.Error("Dgraph Blocks", err, logger.Params{})
		if strings.Contains(err.Error(), "0.000000") {
			return 0, nil
		}
		return 0, err
	}
	if len(r.H) != 1 {
		return 0, errors.New("Something went wrong retrieving max height")
	}
	return r.H[0].Height, nil
}

// LastBlock returnes the last block synced by Bitgodine
func LastBlock() (Block, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), `{
		var(func: has(prev_block)) {
			blocks_height as height
		}
		var() {
			h as max(val(blocks_height))
		}
    b(func: eq(height, val(h))) {
			uid
      id
			height
			version
			prev_block
			merkle_root
      timestamp
      bits
      nonce
    }
	}`)
	if err != nil {
		return Block{}, err
	}

	var r struct{ B []struct{ Block } }
	// var r interface{}
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Block{}, err
	}
	if len(r.B) != 1 {
		return Block{}, errors.New("Something went wrong retrieving last block")
	}
	return r.B[0].Block, nil
}

// StoredBlocks returns an array containing all blocks stored on dgraph
func StoredBlocks() ([]Block, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), `{
		blocks(func: has(prev_block)) {
			height
			id
		}
	}`)
	if err != nil {
		return nil, err
	}

	var r struct{ Blocks []struct{ Block } }
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, err
	}
	var blocks []Block
	for _, b := range r.Blocks {
		blocks = append(blocks, b.Block)
	}

	return blocks, nil
}

// RemoveBlock removes the block specified by its height
func RemoveBlock(block *Block) error {
	return Delete(block.UID)
}

// GetBlockUIDFromHeight returnes the dgraph uid of the block stored at the passed height
func GetBlockUIDFromHeight(height int32) ([]string, error) {
	resp, err := instance.NewReadOnlyTxn().Query(context.Background(), fmt.Sprintf(`{
		block(func: eq(height, %d)) {
			uid
		}
	}`, height))
	if err != nil {
		return nil, err
	}
	var r struct {
		Block []struct {
			UID string `json:"uid,omitempty"`
		}
	}
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, err
	}
	if len(r.Block) == 0 {
		return nil, errors.New("Block not found")
	}
	var uids []string
	for _, b := range r.Block {
		uids = append(uids, b.UID)
	}
	return uids, nil
}
