package bitcoin

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/buffer"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// Block composition to enhance btcutil.Block with other receivers
type Block struct {
	btcutil.Block
}

// Store prepares the block struct and and call StoreBlock to store it
func (b *Block) Store(db kv.DB, height int32) (err error) {
	b.SetHeight(height)
	if height%100 == 0 {
		logger.Info("Parser Blocks", "Block "+strconv.Itoa(int(b.Height())), logger.Params{"hash": b.Hash().String(), "height": b.Height()})
	}
	logger.Debug("Parser Blocks", "Storing block", logger.Params{"hash": b.Hash().String(), "height": height})

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
	err = block.NewService(db, nil).StoreBlock(&blk, transactions)
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

// ExtractBlockFromFile reads and remove magic bytes and size from file and returns Block through btcutil.NewBlockFromBytes
func ExtractBlockFromFile(file *[]uint8) (blk *Block, err error) {
	for len(*file) > 0 && (*file)[0] == 0 {
		*file = (*file)[1:]
	}
	if len(*file) == 0 {
		err = ErrEmptySliceParse
		return
	}
	blockMagic, err := buffer.ReadUint32(file)
	if err != nil {
		return
	}
	switch blockMagic {
	case 0x00:
		err = ErrIncompleteBlockParse
		return
	case 0xd9b4bef9:
		size, e := buffer.ReadUint32(file)
		if e != nil {
			return nil, e
		}
		if size < 80 {
			err = ErrBlockParse
			return
		}
		block, e := buffer.ReadSlice(file, uint(size))
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

// FileKey key for kv stored parsed files counter
const FileKey = "file"

// StoreFileParsed set file stored so far
func StoreFileParsed(db kv.DB, file int) (err error) {
	f := strconv.Itoa(file)
	err = db.Store(FileKey, []byte(f))
	return
}

// GetFileParsed returnes the file parsed so far
func GetFileParsed(db kv.DB) (file int, err error) {
	f, err := db.Read(FileKey)
	if err != nil {
		if errors.Is(err, errorx.ErrKeyNotFound) {
			return 0, nil
		}
		return
	}
	file, err = strconv.Atoi(string(f))
	return
}
