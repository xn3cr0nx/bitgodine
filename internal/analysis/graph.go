package analysis

import (
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
)

// Chunk struct with info on previous analyzed blocks slice
type Chunk struct {
	Range          `json:"range,omitempty"`
	Vulnerabilites Graph `json:"vulnerabilities,omitempty"`
}

// Graph interface to manage analysis result
type Graph interface {
	ExtractPercentages(heuristics.Mask, int32, int32) map[int32][]float64
	ExtractGlobalPercentages(heuristics.Mask, int32, int32) []float64
	subGraph(int32, int32) Graph
	mergeGraphs(...Graph) Graph
	mergeChunks(...Chunk) Chunk
	updateStoredRanges(*badger.Badger, int32, []Chunk) Graph
}

// MaskGraph alias for struct describing blockchain graph based on vulnerabilities mask
type MaskGraph map[int32]map[string]heuristics.Mask

// ExtractPercentages returnes the corresponding map with heuristic percentages for each element in the map (in each block)
func (g MaskGraph) ExtractPercentages(heuristicsList heuristics.Mask, from, to int32) (perc map[int32][]float64) {
	perc = make(map[int32][]float64, to-from+1)
	list := heuristicsList.ToList()
	for i := from; i <= to; i++ {
		perc[i] = make([]float64, len(list))
		for h, heuristic := range list {
			counter := 0
			if len(g[i]) == 0 {
				perc[i][h] = 0
				continue
			}
			for _, v := range g[i] {
				if v.VulnerableMask(heuristic) {
					counter++
				}
			}
			perc[i][h] = float64(counter) / float64(len(g[i]))
		}
	}
	return
}

// ExtractGlobalPercentages returnes the corresponding map with global heuristic percentages for each heuristic
func (g MaskGraph) ExtractGlobalPercentages(heuristicsList heuristics.Mask, from, to int32) (perc []float64) {
	list := heuristicsList.ToList()
	perc = make([]float64, len(list))
	for h, heuristic := range list {
		counter, tot := 0, 0
		for i := from; i <= to; i++ {
			for _, v := range g[i] {
				if v.VulnerableMask(heuristic) {
					counter++
				}
				tot++
			}
		}
		perc[h] = float64(counter) / float64(tot)
	}
	return
}

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

func (g MaskGraph) subGraph(from, to int32) (sub Graph) {
	sub = make(MaskGraph, to-from+1)
	for h, a := range g {
		if h >= from && h <= to {
			sub.(MaskGraph)[h] = a
		}
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
func (g MaskGraph) mergeGraphs(args ...Graph) (merged Graph) {
	merged = g
	if len(args) == 0 {
		return
	}
	for _, arg := range args {
		for height, txs := range arg.(MaskGraph) {
			merged.(MaskGraph)[height] = txs
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
func (g MaskGraph) mergeChunks(args ...Chunk) (merged Chunk) {
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
