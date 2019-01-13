package blocks

import (
	"errors"

	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/pkg/buffer"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// CheckBlock checks if block is correctly initialized just checking hash and height fields have some value
func CheckBlock(b *btcutil.Block) bool {
	return b.Hash() != nil && b.Height() != 0
}

// Walk parses the block and iterates over block's transaction to parse them
func Walk(b *btcutil.Block, v *visitor.BlockchainVisitor, height *uint64, outputItems *map[chainhash.Hash][]visitor.OutputItem) {
	timestamp := b.MsgBlock().Header.Timestamp
	blockItem := (*v).VisitBlockBegin(b, *height)
	for _, tx := range b.Transactions() {
		txs.Walk(tx, v, timestamp, *height, &blockItem, outputItems)
	}
	(*v).VisitBlockEnd(b, *height, blockItem)
}

// Read reads and remove magic bytes and size from slice and returns btcutil.Block through btcutil.NewBlockFromBytes
func Read(slice *[]uint8) (*btcutil.Block, error) {
	for len(*slice) > 0 && (*slice)[0] == 0 {
		*slice = (*slice)[1:]
	}
	if len(*slice) == 0 {
		err := errors.New("Cannot read block from slice")
		logger.Error("Blockchain", err, logger.Params{})
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
		return res, nil
	default:
		err := errors.New("No magic bytes matching")
		logger.Error("Blockchain", err, logger.Params{})
		return nil, err
	}
}
