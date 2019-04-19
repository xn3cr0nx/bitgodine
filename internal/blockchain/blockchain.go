package blockchain

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	mmap "github.com/edsrzf/mmap-go"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Blockchain data structure composed by the memory mapped files in array of mmaps and network conofiguration
type Blockchain struct {
	Maps    []mmap.MMap
	Network chaincfg.Params
	height  int32
}

var blockchain *Blockchain

// Instance singleton pattern return always the same instance of blockchain. In the first time initializes the blockchain
func Instance(network chaincfg.Params) *Blockchain {
	if blockchain == nil {
		blockchain = new(Blockchain)
		blockchain.Network = network
		height := blockchain.Height()
		blockchain.height = height
	}
	return blockchain
}

func (b *Blockchain) Read() error {
	var Maps []mmap.MMap
	netPath := b.Network.Name
	n := 0
	blocksDir := viper.GetString("blocksDir")
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

// Height returnes the height of the last block in the blockchain (currently synced)
func (b *Blockchain) Height() int32 {
	if b.height != 0 {
		return b.height
	}
	height, err := dgraph.LastBlockHeight()
	if err != nil {
		logger.Panic("Blockchain", err, logger.Params{})
	}
	return height
}

// Head returnes the last block in the blockchain
func (b *Blockchain) Head() (blocks.Block, error) {
	last, err := dgraph.LastBlock()
	if err != nil {
		return blocks.Block{}, err
	}
	block, err := last.GenerateBlock()
	if err != nil {
		return blocks.Block{}, err
	}
	return block, nil
}
