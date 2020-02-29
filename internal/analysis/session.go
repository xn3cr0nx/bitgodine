package analysis

import (
	"fmt"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
)

func restorePreviousAnalysis(kv *badger.Badger, from, to, interval int32) (intervals []Chunk) {
	if to-from >= interval {
		upper := upperBoundary(from, interval)
		lower := lowerBoundary(to, interval)
		fmt.Println("restoring in range", upper, lower, interval)
		for i := upper; i < lower; i += interval {
			r, err := kv.Read(fmt.Sprintf("int%d-%d", i, i+interval))
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
		r, err := kv.Read(fmt.Sprintf("int%d-%d", lower, upper))
		if err != nil {
			return
		}
		var analyzed Chunk
		err = encoding.Unmarshal(r, &analyzed)
		if err != nil {
			logger.Error("Analysis", err, logger.Params{})
		}
		analyzed.Vulnerabilites = subGraph(analyzed.Vulnerabilites, from, to)
		intervals = []Chunk{analyzed}
	}
	return
}

// storeRange stores sub chunks of analysis graph based on the interval
func storeRange(kv *badger.Badger, r Range, interval int32, vuln MaskGraph) (err error) {
	upper := upperBoundary(r.From, interval)
	lower := lowerBoundary(r.To, interval)
	if lower-upper >= interval {
		for i := upper; i < lower; i += interval {
			fmt.Println("storing chunk", i, i+interval)
			var analyzed Chunk
			analyzed.From = i
			analyzed.To = i + interval
			analyzed.Vulnerabilites = subGraph(vuln, i, i+interval)
			var a []byte
			a, err = encoding.Marshal(analyzed)
			if err != nil {
				return
			}
			if err = kv.Store(fmt.Sprintf("int%d-%d", i, i+interval), a); err != nil {
				return
			}
		}
	}
	return
}

// updateStoredRange updates sub chunks of analysis graph based on the interval with new analysis
func updateStoredRanges(kv *badger.Badger, interval int32, analyzed []Chunk, vuln MaskGraph) (newVuln MaskGraph) {
	if len(analyzed) == 0 {
		return vuln
	}
	newRange := Range{From: analyzed[0].From, To: analyzed[len(analyzed)-1].To}
	newVuln = make(MaskGraph, newRange.To-newRange.From+1)
	for _, a := range analyzed {
		for i := a.From; i <= a.To; i++ {
			newVuln[i] = make(map[string]heuristics.Mask, len(a.Vulnerabilites[i]))
			for tx, v := range a.Vulnerabilites[i] {
				newVuln[i][tx] = heuristics.MergeMask(v, vuln[i][tx])
			}
		}
	}
	return
}
