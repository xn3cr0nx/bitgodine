package blockchain

import (
	"fmt"
	"testing"

	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
	"github.com/xn3cr0nx/bitgodine_code/internal/clusterizer"
)

func TestInstance(t *testing.T) {
	blockchain1 := Instance(chaincfg.MainNetParams)
	blockchain2 := Instance(chaincfg.MainNetParams)

	// Comparing pointers -> returned the same instance (singleton pattern working)
	assert.Equal(t, blockchain1, blockchain2)
}

func TestWalk(t *testing.T) {
	b := Instance(chaincfg.MainNetParams)
	b.Read()

	cltz := clusterizer.NewClusterizer()
	b.Walk(cltz)
	cltzCount, err := cltz.Done()
	if err != nil {
		logger.Error("Blockchain test", err, logger.Params{})
	}
	fmt.Printf("Clusters: %v\n", cltzCount)
}
