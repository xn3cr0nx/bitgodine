package analysis

import (
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
)

// HeuristicChangeAnalysis analysis output map for heuristics change output
type HeuristicChangeAnalysis map[heuristics.Heuristic]uint32

// OutputGraph alias for struct describing blockchain graph based on heuristics
type OutputGraph map[int32]map[string]HeuristicChangeAnalysis

// ExtractPercentages returnes the corresponding map with heuristic percentages for each element in the map (in each block)
func (g OutputGraph) ExtractPercentages(heuristicsList heuristics.Mask, from, to int32) (perc map[int32][]float64) {
	perc = make(map[int32][]float64, to-from+1)
	list := heuristicsList.ToList()
	for i := from; i <= to; i++ {
		perc[i] = make([]float64, len(list))
		for h, heuristic := range list {
			counter := 0
			for _, v := range g[i] {
				if change, ok := v[heuristic]; ok {
					if change == 0 {
						counter++
					}
				}
				perc[i][h] = float64(counter) / float64(len(g[i]))
			}
		}
	}
	return
}

// ExtractGlobalPercentages returnes the corresponding map with global heuristic percentages for each heuristic
func (g OutputGraph) ExtractGlobalPercentages(heuristicsList heuristics.Mask, from, to int32) (perc []float64) {
	list := heuristicsList.ToList()
	perc = make([]float64, len(list))
	for h, heuristic := range list {
		counter, tot := 0, 0
		for i := from; i <= to; i++ {
			for _, v := range g[i] {
				if change, ok := v[heuristic]; ok {
					// FIXME: here is where I calculated the amount of 0 indexed outputs
					if change == 0 {
						counter++
					}
				}
				tot++
			}
		}
		perc[h] = float64(counter) / float64(tot)
	}
	return
}

func (g OutputGraph) subGraph(from, to int32) (sub Graph) {
	sub = make(OutputGraph, to-from+1)
	for h, a := range g {
		if h >= from && h <= to {
			sub.(OutputGraph)[h] = a
		}
	}
	return
}

// mergeGraphs returnes the union of multiple graphs
func (g OutputGraph) mergeGraphs(args ...Graph) (merged Graph) {
	// merged = make(OutputGraph)
	merged = g
	if len(args) == 0 {
		return
	}
	for _, arg := range args {
		for height, txs := range arg.(OutputGraph) {
			merged.(OutputGraph)[height] = txs
		}
	}
	return
}

// mergeChunks returns the union of multiple chunks
func (g OutputGraph) mergeChunks(args ...Chunk) (merged Chunk) {
	merged = Chunk{
		Vulnerabilites: g,
	}
	if len(args) == 0 {
		return
	}
	min, max := args[0].From, args[0].To
	for _, chunk := range args {
		if chunk.From < min {
			min = chunk.From
		}
		if chunk.To < max {
			max = chunk.To
		}
		merged.Vulnerabilites = merged.Vulnerabilites.mergeGraphs(chunk.Vulnerabilites)
	}
	merged.From, merged.To = min, max
	return
}

// updateStoredRange updates sub chunks of analysis graph based on the interval with new analysis
func (g OutputGraph) updateStoredRanges(kv *badger.Badger, interval int32, analyzed []Chunk) Graph {
	if len(analyzed) == 0 {
		return g
	}
	newRange := Range{From: analyzed[0].From, To: analyzed[len(analyzed)-1].To}
	newGraph := make(OutputGraph, newRange.To-newRange.From+1)
	for _, a := range analyzed {
		for i := a.From; i <= a.To; i++ {
			newGraph[i] = make(map[string]HeuristicChangeAnalysis, len(a.Vulnerabilites.(OutputGraph)[i]))
			for tx, v := range a.Vulnerabilites.(OutputGraph)[i] {
				newGraph[i][tx] = mergeChangeAnalysis(v, g[i][tx])
			}
		}
	}
	return newGraph
}

func mergeChangeAnalysis(a, b HeuristicChangeAnalysis) (new HeuristicChangeAnalysis) {
	new = a
	for h, c := range b {
		a[h] = c
	}
	return
}
