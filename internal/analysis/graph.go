package analysis

import (
	"github.com/xn3cr0nx/bitgodine_parser/pkg/badger"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
)

// Graph interface to manage analysis result
type Graph interface {
	ExtractPercentages(heuristics.Mask, int32, int32) map[int32][]float64
	ExtractGlobalPercentages(heuristics.Mask, int32, int32) []float64
	subGraph(int32, int32) Graph
	mergeGraphs(...Graph) Graph
	mergeChunks(...Chunk) Chunk
	updateStoredRanges(*badger.Badger, int32, []Chunk) Graph
}
