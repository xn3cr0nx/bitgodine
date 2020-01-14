package analysis

import (
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger/kv"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/cache"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
)

func BenchmarkAnalyzeBlocks(t *testing.B) {
	logger.Setup()

	ca, err := cache.NewCache(nil)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	db, err := kv.NewKV(kv.Conf("/home/xn3cr0nx/.bitgodine/badger"), ca, false)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	bdg, err := badger.NewBadger(badger.Conf("/home/xn3cr0nx/.bitgodine/analysis"), false)
	if err != nil {
		logger.Error("Bitgodine", err, logger.Params{})
		os.Exit(-1)
	}

	var vuln map[int32][]byte
	for x := 0; x < t.N; x++ {
		c := echo.New().AcquireContext()
		c.Set("db", db)
		c.Set("cache", ca)
		c.Set("kv", bdg)
		vuln, err = AnalyzeBlocks(&c, 0, 50000, false)
	}
	if len(vuln) == 0 {
		t.Error("failed benchmark for AnalyzeBlocks: ", vuln, err)
	}
}