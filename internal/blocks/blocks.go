package blocks

import (
	"errors"
	// "fmt"
	"math"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/pkg/buffer"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Block composition to enhance btcutil.Block with other receivers
type Block struct {
	btcutil.Block
}

// GenerateBlock converts the Block node struct to a btcsuite Block struct
func GenerateBlock(block *dgraph.Block) (Block, error) {
	prevHash, err := chainhash.NewHashFromStr(block.PrevBlock)
	if err != nil {
		return Block{}, err
	}
	merkleHash, err := chainhash.NewHashFromStr(block.MerkleRoot)
	if err != nil {
		return Block{}, err
	}
	header := wire.NewBlockHeader(block.Version, prevHash, merkleHash, block.Bits, block.Nonce)
	header.Timestamp = block.Time
	msgBlock := wire.NewMsgBlock(header)
	msgBlock.ClearTransactions()
	for _, tx := range block.Transactions {
		t, err := txs.GenerateTransaction(&tx)
		if err != nil {
			return Block{}, err
		}
		if err := msgBlock.AddTransaction(t.MsgTx().Copy()); err != nil {
			return Block{}, err
		}
	}
	b := btcutil.NewBlock(msgBlock)
	return Block{Block: *b}, nil
}

// Store prepares the dgraph block struct and and call StoreBlock to store it in dgraph
func (b *Block) Store() error {
	transactions, err := txs.PrepareTransactions(b.Transactions())
	if err != nil {
		return err
	}
	block := dgraph.Block{
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
	// if err := dgraph.StoreBlock(&block); err != nil {
	if err := dgraph.Store(&block); err != nil {
		return err
	}
	return nil
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
	// fmt.Println("what the slice?", len(*slice))
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
