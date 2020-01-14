package analysis

import (
	"encoding/json"
	"fmt"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
)

// Range wrapper for blocks interval boundaries
type Range struct {
	From int32 `json:"from,omitempty"`
	To   int32 `json:"to,omitempty"`
}

func upperBoundary(n, interval int32) (r int32) {
	diff := n % interval
	if diff == 0 {
		diff = interval
	}
	r = n + (interval - diff)
	return
}

func lowerBoundary(n, interval int32) (r int32) {
	r = n - (n % interval)
	return
}

func updateRange(from, to int32, analyzed []Chunk) (ranges []Range) {
	for i, a := range analyzed {
		if i == 0 {
			if a.From > from {
				ranges[0].To = a.From - 1
			} else {
				ranges[0].To = a.From
			}
		}
		if i == len(analyzed)-1 && a.To < to {
			ranges = append(ranges, Range{a.To + 1, to})
		}
	}
	return
}

func storeRange(kv *badger.Badger, r Range, interval int32, vuln Graph) (err error) {
	upper := upperBoundary(r.From, interval)
	lower := lowerBoundary(r.To, interval)
	if lower-upper >= interval {
		for i := upper; i < lower; i += interval {
			var analyzed Chunk
			analyzed.From = i
			analyzed.To = i + interval
			analyzed.Vulnerabilites = subGraph(vuln, i, i+interval)
			var a []byte
			a, err = json.Marshal(analyzed)
			if err != nil {
				return
			}
			err = kv.Store(fmt.Sprintf("int%d-%d", i, i+interval), a)
		}
	}

	return
}

func mergeGraphs(args ...Graph) (merged Graph) {
	merged = make(Graph)
	for _, arg := range args {
		for height, txs := range arg {
			merged[height] = txs
		}
	}
	return
}

func subGraph(arg Graph, from, to int32) (sub Graph) {
	sub = make(Graph)
	for h, a := range arg {
		sub[h] = a
	}
	return
}

func mergeChunks(args ...Chunk) (merged Chunk) {
	merged = Chunk{
		Vulnerabilites: make(Graph),
	}
	for _, chunk := range args {
		merged.Vulnerabilites = mergeGraphs(merged.Vulnerabilites, chunk.Vulnerabilites)
	}
	return
}
