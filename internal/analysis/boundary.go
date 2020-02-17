package analysis

import (
	"fmt"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
)

// Range wrapper for blocks interval boundaries
type Range struct {
	From int32 `json:"from,omitempty"`
	To   int32 `json:"to,omitempty"`
}

// upperBoundary returnes the nearest upper boundary defined as the
// nearest multiple of the interval above number n
func upperBoundary(n, interval int32) (r int32) {
	diff := n % interval
	if diff == 0 {
		diff = interval
	}
	r = n + (interval - diff)
	return
}

// lowerBoundary returnes the nearest lower boundary defined as the
// nearest multiple of the interval below number n
func lowerBoundary(n, interval int32) (r int32) {
	r = n - (n % interval)
	return
}

// updateRange returns block ranges to be analyzed excluding already analyzed chunks
func updateRange(from, to int32, analyzed []Chunk, force bool) (ranges []Range) {
	ranges = append(ranges, Range{from, to})
	if force {
		return
	}
	for i, a := range analyzed {
		if i == 0 {
			if a.From > from {
				ranges[0].To = a.From - 1
			} else {
				ranges[0].To = a.From
			}
		}
		if i == len(analyzed)-1 && a.To < to {
			ranges = append(ranges, Range{a.To, to})
		}
	}
	return
}

// storeRange stores sub chunks of analysis graph based on the interval
func storeRange(kv *badger.Badger, r Range, interval int32, vuln Graph) (err error) {
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
func updateStoredRanges(kv *badger.Badger, interval int32, analyzed []Chunk, vuln Graph) (newVuln Graph) {
	if len(analyzed) == 0 {
		return vuln
	}
	newRange := Range{From: analyzed[0].From, To: analyzed[len(analyzed)-1].To}
	newVuln = make(Graph, newRange.To-newRange.From+1)
	for _, a := range analyzed {
		for i := a.From; i <= a.To; i++ {
			newVuln[i] = make(map[string]byte, len(a.Vulnerabilites[i]))
			for tx, v := range a.Vulnerabilites[i] {
				newVuln[i][tx] = heuristics.MergeMask(v, vuln[i][tx])
			}
		}
	}
	return
}

// subGraph returnes a graph interval between from and to
func subGraph(g Graph, from, to int32) (sub Graph) {
	sub = make(Graph, to-from+1)
	for h, a := range g {
		if h >= from && h <= to {
			sub[h] = a
		}
	}
	return
}

// mergeGraphs returnes the union of multiple graphs
func mergeGraphs(args ...Graph) (merged Graph) {
	merged = make(Graph)
	for _, arg := range args {
		for height, txs := range arg {
			merged[height] = txs
		}
	}
	return
}

// mergeChunks returns the union of multiple chunks
func mergeChunks(args ...Chunk) (merged Chunk) {
	merged = Chunk{
		Vulnerabilites: make(Graph),
	}
	min, max := args[0].From, args[0].To
	for _, chunk := range args {
		if chunk.From < min {
			min = chunk.From
		}
		if chunk.To < max {
			max = chunk.To
		}
		merged.Vulnerabilites = mergeGraphs(merged.Vulnerabilites, chunk.Vulnerabilites)
	}
	merged.From, merged.To = min, max
	return
}
