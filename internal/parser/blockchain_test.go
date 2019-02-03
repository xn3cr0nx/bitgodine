package parser

import (
	"fmt"
	"testing"

	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
)

func init() {
	logger.Setup()
}

func TestWalk(t *testing.T) {
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
