package blocks

import (
	"errors"
	"math"

	"github.com/btcsuite/btcutil"

	"github.com/xn3cr0nx/bitgodine_code/pkg/buffer"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

type Block struct {
	btcutil.Block
}

// CheckBlock checks if block is correctly initialized just checking hash and height fields have some value
func (b *Block) CheckBlock() bool {
	// return b.Height() != 0 && b.Hash() != nil
	return b != nil && b.Height() == -1 && b.Hash() != nil
}

// CoinbaseValue returnes the value the block should receive from a coinbase transaction based on number of halving happened due to block height
func CoinbaseValue(height int32) int64 {
	return int64(5000000000 / math.Pow(2, float64(height/int32(210000))))
}

// Parse reads and remove magic bytes and size from slice and returns Block through btcutil.NewBlockFromBytes
func Parse(slice *[]uint8) (*Block, error) {
	for len(*slice) > 0 && (*slice)[0] == 0 {
		*slice = (*slice)[1:]
	}
	if len(*slice) == 0 {
		err := errors.New("Cannot read block from slice")
		logger.Info("Blockchain", err.Error(), logger.Params{})
		return nil, err
	}
	blockMagic, err := buffer.ReadUint32(slice)
	if err != nil {
		logger.Error("Blockchain", err, logger.Params{})
		return nil, err
	}
	switch blockMagic {
	case 0x00:
		return nil, errors.New("Incomplete blk file")
	case 0xd9b4bef9:
		size, err := buffer.ReadUint32(slice)
		if err != nil {
			logger.Error("Blockchain", err, logger.Params{})
			return nil, err
		}
		if size < 80 {
			err := errors.New("Cannot parse block")
			logger.Error("Blockchain", err, logger.Params{})
			return nil, err
		}
		block, err := buffer.ReadSlice(slice, uint(size))
		if err != nil {
			logger.Error("Blockchain", err, logger.Params{})
			return nil, err
		}
		res, err := btcutil.NewBlockFromBytes(block)
		if err != nil {
			logger.Error("Blockchain", err, logger.Params{})
			return nil, err
		}
		blk := &Block{Block: *res}
		return blk, nil
	default:
		err := errors.New("No magic bytes matching")
		logger.Error("Blockchain", err, logger.Params{})
		return nil, err
	}
}
