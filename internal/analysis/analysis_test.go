package analysis

import (
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine/pkg/badger"
	badgerStorage "github.com/xn3cr0nx/bitgodine/pkg/badger/storage"
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

	db, err := badgerStorage.NewKV(badgerStorage.Conf("/home/xn3cr0nx/.bitgodine/badger"), c, false)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	bdg, err := badger.NewBadger(badger.Conf("/home/xn3cr0nx/.bitgodine/analysis"), false)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	for x := 0; x < t.N; x++ {
		c := echo.New().AcquireContext()
		c.Set("db", db)
		c.Set("kv", bdg)
		err = AnalyzeBlocks(&c, 0, 120000, heuristics.FromListToMask(heuristics.List()), "applicability", "", "", false)
	}
}
