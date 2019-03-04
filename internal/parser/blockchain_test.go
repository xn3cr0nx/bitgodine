package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
	"gopkg.in/go-playground/assert.v1"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
)

func init() {
	logger.Setup()
}

func TestWalk(t *testing.T) {
	wd, err := os.Getwd()
	assert.Equal(t, err, nil)
	db, _ := db.Instance(&db.Config{Dir: filepath.Join(wd, "..", ".."), Name: "leveldb", Net: wire.MainNet})
	fmt.Println("leveldb", db)

	dgo := dgraph.Instance(&dgraph.Config{Host: "localhost", Port: 9080})
	fmt.Println("dgraph", dgo)
	if err := dgraph.Setup(dgo); err != nil {
		logger.Error("Blockchain test", err, logger.Params{})
	}

	b := blockchain.Instance(chaincfg.MainNetParams)
	b.Read()

	cltz := visitor.NewClusterizer()
	Walk(b, cltz)
	cltzCount, err := cltz.Done()
	if err != nil {
		logger.Error("Blockchain test", err, logger.Params{})
	}
	fmt.Printf("Clusters: %v\n", cltzCount)
}
