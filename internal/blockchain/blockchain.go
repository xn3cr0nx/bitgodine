package blockchain

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	mmap "github.com/edsrzf/mmap-go"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// Blockchain data structure composed by the memory mapped files in array of mmaps and network conofiguration
type Blockchain struct {
	Maps    []mmap.MMap
	Network chaincfg.Params
	db      storage.DB
	height  int32
}

var blockchain *Blockchain

// Instance singleton pattern return always the same instance of blockchain. In the first time initializes the blockchain
func Instance(db storage.DB, network chaincfg.Params) *Blockchain {
	if blockchain == nil {
		blockchain = new(Blockchain)
		blockchain.Network = network
		blockchain.db = db
		height, err := blockchain.Height()
		if err != nil {
			logger.Panic("Blockchain", err, logger.Params{})
		}
		blockchain.height = height
	}
	return blockchain
}

func (b *Blockchain) Read(path string) error {
	var Maps []mmap.MMap
	netPath := b.Network.Name
	if path == "" {
		path = viper.GetString("blocksDir")
	}
	if b.Network.Name == "mainnet" {
		netPath = ""
	}
	if path == "" {
		hd, err := homedir.Dir()
		if err != nil {
			return err
		}
		path = hd
	}
	path = filepath.Join(path, ".bitcoin", netPath, "blocks")

	for n := 0; ; n++ {
		f, err := os.OpenFile(filepath.Join(path, fmt.Sprintf("blk%05d.dat", n)), os.O_RDWR, 0644)
		defer f.Close()
		if err != nil {
			logger.Info("Blockchain", err.Error(), logger.Params{})
			break
		}

		m, err := mmap.Map(f, 2, 0)
		if err != nil {
			return err
		}
		Maps = append(Maps, m)
	}
	b.Maps = Maps
	return nil
}

// Height returnes the height of the last block in the blockchain (currently synced)
func (b *Blockchain) Height() (h int32, err error) {
	h, err = b.db.GetLastBlockHeight()
	return
}

// Head returnes the last block in the blockchain
func (b *Blockchain) Head() (last models.Block, err error) {
	last, err = b.db.LastBlock()
	return
}
