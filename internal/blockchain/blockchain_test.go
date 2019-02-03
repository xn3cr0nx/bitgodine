package blockchain

import (
	"testing"

	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.Setup()
}

func TestInstance(t *testing.T) {
	blockchain1 := Instance(chaincfg.MainNetParams)
	blockchain2 := Instance(chaincfg.MainNetParams)

	// Comparing pointers -> returned the same instance (singleton pattern working)
	assert.Equal(t, blockchain1, blockchain2)
}
