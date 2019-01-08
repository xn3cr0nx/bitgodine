package blockchain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine/pkg/buffer"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"

	"github.com/edsrzf/mmap-go"
	dir "github.com/mitchellh/go-homedir"
)

type Blockchain struct {
	Maps    []mmap.MMap
	Network chaincfg.Params
}

var blockchain *Blockchain

func Instance(network chaincfg.Params) *Blockchain {
	if blockchain == nil {
		blockchain = new(Blockchain)
		blockchain.Network = network
	}
	return blockchain
}

func (b *Blockchain) Read() error {
	var Maps []mmap.MMap
	netPath := b.Network.Name
	n := 0
	blocksDir, _ := dir.Dir()
	if b.Network.Name == "mainnet" {
		netPath = ""
	}
	blocksDir = filepath.Join(blocksDir, ".bitcoin", netPath, "blocks")
	fmt.Printf("%v\n", blocksDir)

	for {
		f, err := os.OpenFile(filepath.Join(blocksDir, fmt.Sprintf("blk%05d.dat", n)), os.O_RDWR, 0644)
		defer f.Close()
		if err != nil {
			break
		}

		n = n + 1
		m, err := mmap.Map(f, 2, 0)
		if err != nil {
			logger.Panic("Mapping file to memory", err, logger.Params{})
			return err
		}
		Maps = append(Maps, m)
	}

	b.Maps = Maps
	return nil
}

// func (b *Blockchain) Walk(v visitor.BlockchainVisitor) (uint64, *chainhash.Hash, map[chainhash.Hash][]visitor.OutputItem) {
// 	var skipped map[chainhash.Hash]btcutil.Block
// 	var outputItems map[chainhash.Hash][]visitor.OutputItem
// 	goalPrevHash, _ := chainhash.NewHash(make([]byte, 32))
// 	var lastBlock btcutil.Block
// 	height := uint64(0)

// 	for k, value := range b.Maps {
// 		logger.Info("Blockchain", "Parsing the blockchain", logger.Params{"file": fmt.Sprintf("%v/%v", k, len(b.Maps)-1)})
// 		b.WalkSlice(value, &goalPrevHash, &lastBlock, &height, &skipped, &outputItems, v)
// 	}

// 	return height, goalPrevHash, outputItems
// }

// func (b *Blockchain) WalkSlice(slice []uint8, goalPrevHash chainhash.Hash, lastBlock btcutil.Block, height *uint64, skipped map[chainhash.Hash]btcutil.Block, outputItems map[chainhash.Hash][]visitor.OutputItem, v visitor.BlockchainVisitor) {
// 	for len(slice) > 0 {
// 		if _, ok := skipped[goalPrevHash]; ok {
// 			lastBlock.Walk(visitor, &height, outputItems)
// 		}
// 	}
// }

func ReadBlock(slice *[]uint8) (*btcutil.Block, error) {
	for len(*slice) > 0 && (*slice)[0] == 0 {
		*slice = (*slice)[1:]
	}
	if len(*slice) == 0 {
		return nil, errors.New("Cannot read block from slice")
	}
	blockMagic, err := buffer.ReadUint32(slice)
	if err != nil {
		return nil, errors.New("Cannot read block magic")
	}
	switch blockMagic {
	case 0x00:
		return nil, errors.New("Incomplete blk file")
	case 0xd9b4bef9:
		size, err := buffer.ReadUint32(slice)
		if err != nil {
			return nil, errors.New("Cannot read block size")
		}
		if size < 80 {
			return nil, errors.New("Cannot parse block")
		}
		block, err := buffer.ReadSlice(slice, uint(size))
		if err != nil {
			return nil, errors.New("Cannot parse block")
		}
		res, err := btcutil.NewBlockFromBytes(block)
		if err != nil {
			return nil, errors.New("Cannot parse block")
		}
		return res, nil
	default:
		return nil, errors.New("No magic bytes matching")
	}
}
