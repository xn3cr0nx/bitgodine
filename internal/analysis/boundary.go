package analysis

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

// subGraph returnes a graph interval between from and to
func subGraph(g MaskGraph, from, to int32) (sub MaskGraph) {
	sub = make(MaskGraph, to-from+1)
	for h, a := range g {
		if h >= from && h <= to {
			sub[h] = a
		}
	}
	return
}

// mergeGraphs returnes the union of multiple graphs
func mergeGraphs(args ...MaskGraph) (merged MaskGraph) {
	merged = make(MaskGraph)
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
		Vulnerabilites: make(MaskGraph),
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
