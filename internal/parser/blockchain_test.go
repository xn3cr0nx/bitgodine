package parser

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
	"gopkg.in/go-playground/assert.v1"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
)

func init() {
	logger.Setup()
}

func TestWalk(t *testing.T) {
	hd, err := homedir.Dir()
	assert.Equal(t, err, nil)
	DbConf := &db.Config{
		Dir: filepath.Join(hd, ".bitgodine", "badger"),
	}
	db, err := db.Instance(DbConf)
	fmt.Println("badger", db)

	dgo := dgraph.Instance(&dgraph.Config{Host: "localhost", Port: 9080})
	fmt.Println("dgraph", dgo)
	if err := dgraph.Setup(dgo); err != nil {
		logger.Error("Blockchain test", err, logger.Params{})
	}

	b := blockchain.Instance(chaincfg.MainNetParams)
	b.Read()

	cltz := visitor.NewClusterizer()
	Walk(b, cltz, nil, nil)
	cltzCount, err := cltz.Done()
	if err != nil {
		logger.Error("Blockchain test", err, logger.Params{})
	}
	fmt.Printf("Clusters: %v\n", cltzCount)
}
