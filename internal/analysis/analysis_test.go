package analysis

import (
	"os"
	"testing"

	"github.com/xn3cr0nx/bitgodine/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv/badger"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

func BenchmarkAnalyzeBlocks(t *testing.B) {
	logger.Setup()

	c, err := cache.NewCache(nil)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	bdg, err := badger.NewBadger(badger.Conf("/home/xn3cr0nx/.bitgodine/analysis"), false)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}
	db, err := badger.NewKV(bdg, c)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	service := NewService(db, nil)
	for x := 0; x < t.N; x++ {
		err = service.AnalyzeBlocks(0, 120000, heuristics.FromListToMask(heuristics.List()), "applicability", "", "", false)
	}
}
