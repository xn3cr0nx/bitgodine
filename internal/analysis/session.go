package analysis

import (
	"fmt"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
)

func restorePreviousAnalysis(kv *badger.Badger, from, to, interval int32, analysisType string) (intervals []Chunk) {
	if to-from >= interval {
		upper := upperBoundary(from, interval)
		lower := lowerBoundary(to, interval)
		fmt.Println("restoring in range", upper, lower, interval)
		for i := upper; i < lower; i += interval {
			r, err := kv.Read(fmt.Sprintf(analysisType+"%d-%d", i, i+interval))
			fmt.Println("read range", i, i+interval, err)
			if err != nil {
				break
			}
			var analyzed Chunk
			err = encoding.Unmarshal(r, &analyzed)
			if err != nil {
				logger.Error("Analysis", err, logger.Params{})
				break
			}
			intervals = append(intervals, analyzed)
		}
	} else {
		lower := lowerBoundary(from, interval)
		upper := upperBoundary(to, interval)
		r, err := kv.Read(fmt.Sprintf(analysisType+"%d-%d", lower, upper))
		if err != nil {
			return
		}
		var analyzed Chunk
		err = encoding.Unmarshal(r, &analyzed)
		if err != nil {
			logger.Error("Analysis", err, logger.Params{})
		}
		analyzed.Vulnerabilites = analyzed.Vulnerabilites.subGraph(from, to)
		intervals = []Chunk{analyzed}
	}
	return
}

// storeRange stores sub chunks of analysis graph based on the interval
func storeRange(kv *badger.Badger, r Range, interval int32, vuln Graph, analysisType string) (err error) {
	upper := upperBoundary(r.From, interval)
	lower := lowerBoundary(r.To, interval)

	if lower-upper >= interval {
		for i := upper; i < lower; i += interval {
			fmt.Println("storing chunk", i, i+interval)
			var analyzed Chunk
			analyzed.From = i
			analyzed.To = i + interval
			analyzed.Vulnerabilites = vuln.subGraph(i, i+interval)
			var a []byte
			a, err = encoding.Marshal(analyzed)
			if err != nil {
				return
			}
			if err = kv.Store(fmt.Sprintf(analysisType+"%d-%d", i, i+interval), a); err != nil {
				return
			}
		}
	}
	return
}
