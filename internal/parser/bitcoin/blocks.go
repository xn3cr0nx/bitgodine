package bitcoin

import (
	"fmt"
	"math"
	"strconv"

	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/buffer"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// Block composition to enhance btcutil.Block with other receivers
type Block struct {
	btcutil.Block
}

// Store prepares the block struct and and call StoreBlock to store it
func (b *Block) Store(db storage.DB, height *int32) (err error) {
	b.SetHeight(*height)
	if *height%100 == 0 {
		logger.Info("Parser Blocks", "Block "+strconv.Itoa(int(b.Height())), logger.Params{"hash": b.Hash().String(), "height": b.Height()})
	}
	logger.Debug("Parser Blocks", "Storing block", logger.Params{"hash": b.Hash().String(), "height": *height})

	transactions, err := PrepareTransactions(db, b.Transactions())
	if err != nil {
		return
	}
	status := []tx.Status{
		{
			Confirmed:   true,
			BlockHeight: b.Height(),
			BlockHash:   b.Hash().String(),
			BlockTime:   b.MsgBlock().Header.Timestamp,
		},
	}
	txsRefs := make([]string, len(transactions))
	for i, transaction := range transactions {
		// TODO: check this assignment, something is wrong here, stored object doesn't have status
		transaction.Status = status
		txsRefs[i] = transaction.TxID
	}

	size, err := b.Bytes()
	if err != nil {
		return
	}
	weight, err := b.BytesNoWitness()
	if err != nil {
		return
	}
	blk := block.Block{
		ID:                b.Hash().String(),
		Height:            b.Height(),
		Version:           b.MsgBlock().Header.Version,
		Timestamp:         b.MsgBlock().Header.Timestamp,
		Bits:              b.MsgBlock().Header.Bits,
		Nonce:             b.MsgBlock().Header.Nonce,
		MerkleRoot:        b.MsgBlock().Header.MerkleRoot.String(),
		Transactions:      txsRefs,
		TxCount:           len(txsRefs),
		Size:              len(size),
		Weight:            len(weight),
		Previousblockhash: b.MsgBlock().Header.PrevBlock.String(),
	}
	err = block.StoreBlock(db, &blk, transactions)
	return
}

// CheckBlock checks if block is correctly initialized just checking hash and height fields have some value
func (b *Block) CheckBlock() bool {
	return b != nil && b.MsgBlock() != nil && b.Hash() != nil
}

// CoinbaseValue returns the value the block should receive from a coinbase transaction based on number of halving happened due to block height
func CoinbaseValue(height int32) int64 {
	return int64(5000000000 / math.Pow(2, float64(height/int32(210000))))
}

// ExtractBlockFromSlice reads and remove magic bytes and size from slice and returns Block through btcutil.NewBlockFromBytes
func ExtractBlockFromSlice(slice *[]uint8) (blk *Block, err error) {
	for len(*slice) > 0 && (*slice)[0] == 0 {
		*slice = (*slice)[1:]
	}
	if len(*slice) == 0 {
		err = ErrEmptySliceParse
		return
	}
	blockMagic, err := buffer.ReadUint32(slice)
	if err != nil {
		return
	}
	switch blockMagic {
	case 0x00:
		err = ErrIncompleteBlockParse
		return
	case 0xd9b4bef9:
		size, e := buffer.ReadUint32(slice)
		if e != nil {
			return nil, e
		}
		if size < 80 {
			err = ErrBlockParse
			return
		}
		block, e := buffer.ReadSlice(slice, uint(size))
		if e != nil {
			return nil, e
		}
		res, e := btcutil.NewBlockFromBytes(block)
		if e != nil {
			err = fmt.Errorf("%s: %w", ErrBlockFromBytes.Error(), e)
			return
		}
		blk = &Block{Block: *res}
		return
	default:
		err = ErrMagicBytesMatching
		return
	}
}
