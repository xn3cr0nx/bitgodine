package blockchain

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	mmap "github.com/edsrzf/mmap-go"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Blockchain data structure composed by the memory mapped files in array of mmaps and network conofiguration
type Blockchain struct {
	Maps    []mmap.MMap
	Network chaincfg.Params
}

var blockchain *Blockchain

// Instance singleton pattern return always the same instance of blockchain. In the first time initializes the blockchain
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
