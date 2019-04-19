package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
)

// Block represent the dgraph node containing essential block info
type Block struct {
	UID          string        `json:"uid,omitempty"`
	Hash         string        `json:"hash,omitempty"`
	Height       int32         `json:"height,omitempty"`
	PrevBlock    string        `json:"prev_block,omitempty"`
	Time         time.Time     `json:"time,omitempty"`
	Transactions []Transaction `json:"transactions,omitempty"`
	Version      int32         `json:"version,omitempty"`
	MerkleRoot   string        `json:"merkle_root,omitempty"`
	Bits         uint32        `json:"bits,omitempty"`
	Nonce        uint32        `json:"nonce,omitempty"`
}

// GenerateBlock converts the Block node struct to a btcsuite Block struct
func (block *Block) GenerateBlock() (blocks.Block, error) {
	prevHash, err := chainhash.NewHashFromStr(block.PrevBlock)
	if err != nil {
		return blocks.Block{}, err
	}
	merkleHash, err := chainhash.NewHashFromStr(block.MerkleRoot)
	if err != nil {
		return blocks.Block{}, err
	}
	header := wire.NewBlockHeader(block.Version, prevHash, merkleHash, block.Bits, block.Nonce)
	msgBlock := wire.NewMsgBlock(header)
	b := btcutil.NewBlock(msgBlock)
	return blocks.Block{Block: *b}, nil
}

// GetBlockHashFromHeight returnes the hash of the block retrieving it based on its height
func GetBlockHashFromHeight(height int32) (string, error) {
	resp, err := instance.NewTxn().Query(context.Background(), fmt.Sprintf(`{
		q(func: eq(height, %d), first: 1) {
			block
		}
	}`, height))
	if err != nil {
		return "", err
	}
	var r Resp
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", err
	}
	if len(r.Q) == 0 {
		return "", errors.New("No address occurences")
	}
	return r.Q[0].Block, nil
}

// func StoreBlock(hash, prev string, height int32, timestamp time.Time, txs []*btcutil.Tx) error {
func StoreBlock(b *blocks.Block) error {
	transactions, err := PrepareTransactions(b.Transactions(), b.Height())
	if err != nil {
		return err
	}
	node := Block{
		Hash:         b.Hash().String(),
		PrevBlock:    b.MsgBlock().Header.PrevBlock.String(),
		Height:       b.Height(),
		Time:         b.MsgBlock().Header.Timestamp,
		Transactions: transactions,
		Version:      b.MsgBlock().Header.Version,
		MerkleRoot:   b.MsgBlock().Header.MerkleRoot.String(),
		Bits:         b.MsgBlock().Header.Bits,
		Nonce:        b.MsgBlock().Header.Nonce,
	}
	out, err := json.Marshal(node)
	if err != nil {
		return err
	}
	_, err = instance.NewTxn().Mutate(context.Background(), &api.Mutation{SetJson: out, CommitNow: true})
	if err != nil {
		return err
	}

	return nil
}

// LastBlockHeight returnes the height of the last block synced by Bitgodine
func LastBlockHeight() (int32, error) {
	resp, err := instance.NewTxn().Query(context.Background(), `{
		var(func: has(hash)) {
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
	resp, err := instance.NewTxn().Query(context.Background(), `{
		var(func: has(hash)) {
			blocks_height as height
		}
		var() {
			h as max(val(blocks_height))
		}
    r(func: eq(height, val(h))) {
      hash
      height
      version
			prev_block
			merkle_root
      time
      bits
      nonce
    }
	}`)
	if err != nil {
		return Block{}, err
	}

	var r struct{ H []struct{ Block } }
	// var r interface{}
	if err := json.Unmarshal(resp.GetJson(), &r); err != nil {
		return Block{}, err
	}
	if len(r.H) != 1 {
		return Block{}, errors.New("Something went wrong retrieving max height")
	}
	return r.H[0].Block, nil
}
