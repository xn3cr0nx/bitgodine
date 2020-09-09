package blocks

import (
	"errors"

	// "fmt"
	"math"

	"github.com/btcsuite/btcutil"

	"github.com/xn3cr0nx/bitgodine/internal/storage"
	txs "github.com/xn3cr0nx/bitgodine/internal/transactions"
	"github.com/xn3cr0nx/bitgodine/pkg/buffer"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// Block composition to enhance btcutil.Block with other receivers
type Block struct {
	btcutil.Block
}

// Store prepares the dgraph block struct and and call StoreBlock to store it in dgraph
func (b *Block) Store(db storage.DB) (err error) {
	transactions, err := txs.PrepareTransactions(db, b.Transactions())
	if err != nil {
		return
	}
	status := []models.Status{
		models.Status{
			Confirmed:   true,
			BlockHeight: b.Height(),
			BlockHash:   b.Hash().String(),
			BlockTime:   b.MsgBlock().Header.Timestamp,
		},
	}
	txsRefs := make([]string, len(transactions))
	for i, tx := range transactions {
		// TODO: check this assignment, something is wrong here, stored object doesn't have status
		tx.Status = status
		txsRefs[i] = tx.TxID
	}

	size, err := b.Bytes()
	if err != nil {
		return
	}
	weight, err := b.BytesNoWitness()
	if err != nil {
		return
	}
	block := models.Block{
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
	err = db.StoreBlock(&block, transactions)
	return
}

// CheckBlock checks if block is correctly initialized just checking hash and height fields have some value
func (b *Block) CheckBlock() bool {
	return b != nil && b.MsgBlock() != nil && b.Hash() != nil
}

// CoinbaseValue returnes the value the block should receive from a coinbase transaction based on number of halving happened due to block height
func CoinbaseValue(height int32) int64 {
	return int64(5000000000 / math.Pow(2, float64(height/int32(210000))))
}

// Parse reads and remove magic bytes and size from slice and returns Block through btcutil.NewBlockFromBytes
func Parse(slice *[]uint8) (blk *Block, err error) {
	for len(*slice) > 0 && (*slice)[0] == 0 {
		*slice = (*slice)[1:]
	}
	if len(*slice) == 0 {
		err = errors.New("Cannot read block from slice")
		logger.Info("Blockchain", err.Error(), logger.Params{})
		return
	}
	blockMagic, err := buffer.ReadUint32(slice)
	if err != nil {
		logger.Error("Blockchain", err, logger.Params{})
		return
	}
	switch blockMagic {
	case 0x00:
		err = errors.New("Incomplete blk file")
		return
	case 0xd9b4bef9:
		size, e := buffer.ReadUint32(slice)
		if e != nil {
			logger.Error("Blockchain", err, logger.Params{})
			return nil, e
		}
		if size < 80 {
			err = errors.New("Cannot parse block")
			logger.Error("Blockchain", err, logger.Params{})
			return
		}
		block, e := buffer.ReadSlice(slice, uint(size))
		if e != nil {
			logger.Error("Blockchain", err, logger.Params{})
			return nil, e
		}
		res, e := btcutil.NewBlockFromBytes(block)
		if e != nil {
			logger.Error("Blockchain", err, logger.Params{})
			return nil, e
		}
		blk = &Block{Block: *res}
		return
	default:
		err = errors.New("No magic bytes matching")
		logger.Error("Blockchain", err, logger.Params{})
		return
	}
}
