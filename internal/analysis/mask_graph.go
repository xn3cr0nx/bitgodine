package analysis

import (
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
)

// MaskGraph alias for struct describing blockchain graph based on vulnerabilities mask
type MaskGraph map[int32]map[string]heuristics.Mask

// ExtractPercentages returnes the corresponding map with heuristic percentages for each element in the map (in each block)
func (g MaskGraph) ExtractPercentages(heuristicsList heuristics.Mask, from, to int32) (perc map[int32][]float64) {
	perc = make(map[int32][]float64, to-from+1)
	list := heuristicsList.ToList()
	for i := from; i <= to; i++ {
		perc[i] = make([]float64, len(list))
		for h, heuristic := range list {
			counter, tot := 0, 0
			if len(g[i]) == 0 {
				perc[i][h] = 0
				continue
			}
			for _, v := range g[i] {
				if !v.IsCoinbase() {
					if v.VulnerableMask(heuristic) {
						counter++
					}
					tot++
				}
			}
			perc[i][h] = float64(counter) / float64(tot)
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
				if !v.IsCoinbase() {
					if v.VulnerableMask(heuristic) {
						counter++
					}
					tot++
				}
			}
		}
		perc[h] = float64(counter) / float64(tot)
	}
	return
}

// ExtractGlobalOffByOneBug extraction based on output graph mock
func (g MaskGraph) ExtractGlobalOffByOneBug(heuristicsList heuristics.Mask, from, to int32) (perc []float64) {
	return
}

// ExtractOffByOneBug extraction based on output graph mock
func (g MaskGraph) ExtractOffByOneBug(heuristicsList heuristics.Mask, from, to int32) (perc map[int32][]float64) {
	return
}

// ExtractGlobalSecureBasisPerc extraction based on output graph mock
func (g MaskGraph) ExtractGlobalSecureBasisPerc(heuristicsList heuristics.Mask, from, to int32) (perc []float64) {
	return
}

// ExtractCombinationPercentages extraction based on output graph mock
func (g MaskGraph) ExtractCombinationPercentages(heuristicsList heuristics.Mask, from, to int32) (perc map[string]float64) {
	return
}

// ExtractGlobalFullMajorityVotingPerc extraction based on output graph mock
func (g MaskGraph) ExtractGlobalFullMajorityVotingPerc(heuristicsList heuristics.Mask, from, to int32) (perc map[string]float64) {
	return
}

// ExtractGlobalMajorityVotingPerc extraction based on output graph mock
func (g MaskGraph) ExtractGlobalMajorityVotingPerc(heuristicsList heuristics.Mask, from, to int32) (perc map[string]float64) {
	return
}

// ExtractGlobalStricMajorityVotingPerc extraction based on output graph mock
func (g MaskGraph) ExtractGlobalStricMajorityVotingPerc(heuristicsList heuristics.Mask, from, to int32) (perc map[string]float64) {
	return
}

// MajorityFullAnalysis extraction based on output graph mock
func (g MaskGraph) MajorityFullAnalysis(heuristicsList heuristics.Mask, from, to int32, reducing ...heuristics.Heuristic) AnalysisSet {
	return AnalysisSet{}
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

// updateStoredRange updates sub chunks of analysis graph based on the interval with new analysis
func (g MaskGraph) updateStoredRanges(kv *badger.Badger, interval int32, analyzed []Chunk) Graph {
	if len(analyzed) == 0 {
		return g
	}
	newRange := Range{From: analyzed[0].From, To: analyzed[len(analyzed)-1].To}
	newGraph := make(MaskGraph, newRange.To-newRange.From+1)
	for _, a := range analyzed {
		for i := a.From; i <= a.To; i++ {
			newGraph[i] = make(map[string]heuristics.Mask, len(a.Vulnerabilites.(MaskGraph)[i]))
			for tx, v := range a.Vulnerabilites.(MaskGraph)[i] {
				newGraph[i][tx] = heuristics.MergeMasks(v, g[i][tx])
			}
		}
	}
	return newGraph
}
