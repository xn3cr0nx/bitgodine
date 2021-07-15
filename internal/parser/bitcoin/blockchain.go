package bitcoin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	mmap "github.com/edsrzf/mmap-go"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/block"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// Blockchain data structure composed by the memory mapped files in array of mmaps and network conofiguration
type Blockchain struct {
	Maps    []mmap.MMap
	Network chaincfg.Params
	db      kv.DB
	height  int32
}

var blockchain *Blockchain

// NewBlockchain singleton pattern return always the same instance of blockchain. In the first time initializes the blockchain
func NewBlockchain(db kv.DB, network chaincfg.Params) *Blockchain {
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

func (b *Blockchain) Read(path string, from int) error {
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
	path = filepath.Join(path, netPath, "blocks")

	for n := from; ; n++ {
		f, err := os.OpenFile(filepath.Join(path, fmt.Sprintf("blk%05d.dat", n)), os.O_RDWR, 0644)
		defer f.Close()
		if err != nil {
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

// Height returns the height of the last block in the blockchain (currently synced)
func (b *Blockchain) Height() (h int32, err error) {
	h, err = block.NewService(b.db, nil).ReadHeight()
	return
}

// Head returns the last block in the blockchain
func (b *Blockchain) Head() (last block.Block, err error) {
	h, err := b.Height()
	if err != nil {
		return
	}

	last, err = block.NewService(b.db, nil).ReadFromHeight(h)
	return
}
