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

func updateRange(from, to int32, analyzed []Analyzed) (ranges []Range) {
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

func storeRange(kv *badger.Badger, r Range, interval int32, vuln map[int32][]byte) (err error) {
	upper := upperBoundary(r.From, interval)
	lower := lowerBoundary(r.To, interval)
	if lower-upper >= interval {
		for i := upper; i < lower; i += interval {
			var analyzed Analyzed
			analyzed.From = i
			analyzed.To = i + interval
			analyzed.Vulnerabilites = subMap(vuln, i, i+interval)
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

func mergeMaps(args ...map[int32][]byte) (merged map[int32][]byte) {
	merged = make(map[int32][]byte)
	for _, arg := range args {
		for height, perc := range arg {
			merged[height] = perc
		}
	}
	return
}

func subMap(arg map[int32][]byte, from, to int32) (sub map[int32][]byte) {
	sub = make(map[int32][]byte)
	for h, a := range arg {
		sub[h] = a
	}
	return
}
