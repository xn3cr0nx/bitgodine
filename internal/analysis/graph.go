package analysis

import (
	"github.com/xn3cr0nx/bitgodine/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine/internal/storage/badger"
)

type AnalysisSet struct {
	LocalPercentages map[byte]float64
	Percentages      map[byte]float64
	LocalCounters    map[byte]float64
	Counters         map[byte]float64
	Combinations     map[byte]float64
}

// Graph interface to manage analysis result
type Graph interface {
	subGraph(int32, int32) Graph
	mergeGraphs(...Graph) Graph
	mergeChunks(...Chunk) Chunk
	updateStoredRanges(*badger.Badger, int32, []Chunk) Graph
	ExtractPercentages(heuristics.Mask, int32, int32) map[int32][]float64
	ExtractGlobalPercentages(heuristics.Mask, int32, int32) []float64
	ExtractOffByOneBug(heuristics.Mask, int32, int32) map[int32][]float64
	ExtractGlobalOffByOneBug(heuristics.Mask, int32, int32) []float64
	ExtractGlobalSecureBasisPerc(heuristics.Mask, int32, int32) []float64
	ExtractCombinationPercentages(heuristics.Mask, int32, int32) map[string]float64
	ExtractGlobalFullMajorityVotingPerc(heuristics.Mask, int32, int32) map[string]float64
	ExtractGlobalMajorityVotingPerc(heuristics.Mask, int32, int32) map[string]float64
	ExtractGlobalStricMajorityVotingPerc(heuristics.Mask, int32, int32) map[string]float64
	MajorityFullAnalysis(heuristics.Mask, int32, int32, ...heuristics.Heuristic) AnalysisSet
}
