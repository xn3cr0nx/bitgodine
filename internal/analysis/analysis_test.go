package analysis

import (
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger/kv"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/cache"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
)

func BenchmarkAnalyzeBlocks(t *testing.B) {
	logger.Setup()

	c, err := cache.NewCache(nil)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	db, err := kv.NewKV(kv.Conf("/home/xn3cr0nx/.bitgodine/badger"), c, false)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	bdg, err := badger.NewBadger(badger.Conf("/home/xn3cr0nx/.bitgodine/analysis"), false)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	var vuln map[int32]map[string]byte
	for x := 0; x < t.N; x++ {
		c := echo.New().AcquireContext()
		c.Set("db", db)
		c.Set("kv", bdg)
		vuln, err = AnalyzeBlocks(&c, 0, 120000, heuristics.List(), false, false)
	}
	if len(vuln) == 0 {
		t.Error("failed benchmark for AnalyzeBlocks: ", vuln, err)
	}
}
