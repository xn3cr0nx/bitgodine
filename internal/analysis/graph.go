package analysis

import "github.com/xn3cr0nx/bitgodine_server/internal/heuristics"

// Graph interface to manage analysis result
type Graph interface {
	ExtractPercentages(heuristics.Mask, int32, int32) map[int32][]float64
	ExtractGlobalPercentages(heuristics.Mask, int32, int32) []float64
}

// MaskGraph alias for struct describing blockchain graph based on vulnerabilities mask
type MaskGraph map[int32]map[string]heuristics.Mask

// Chunk struct with info on previous analyzed blocks slice
type Chunk struct {
	Range          `json:"range,omitempty"`
	Vulnerabilites MaskGraph `json:"vulnerabilities,omitempty"`
}

// HeuristicGraph alias for struct describing blockchain graph based on heuristics
type HeuristicGraph map[int32]map[string]HeuristicChangeAnalysis

// ExtractPercentages returnes the corresponding map with heuristic percentages for each element in the map (in each block)
func (data MaskGraph) ExtractPercentages(heuristicsList heuristics.Mask, from, to int32) (perc map[int32][]float64) {
	perc = make(map[int32][]float64, to-from+1)
	list := heuristicsList.ToList()
	for i := from; i <= to; i++ {
		perc[i] = make([]float64, len(list))
		for h, heuristic := range list {
			counter := 0
			if len(data[i]) == 0 {
				perc[i][h] = 0
				continue
			}
			for _, v := range data[i] {
				if v.VulnerableMask(heuristic) {
					counter++
				}
			}
			perc[i][h] = float64(counter) / float64(len(data[i]))
		}
	}
	return
}

// ExtractGlobalPercentages returnes the corresponding map with global heuristic percentages for each heuristic
func (data MaskGraph) ExtractGlobalPercentages(heuristicsList heuristics.Mask, from, to int32) (perc []float64) {
	list := heuristicsList.ToList()
	perc = make([]float64, len(list))
	for h, heuristic := range list {
		counter, tot := 0, 0
		for i := from; i <= to; i++ {
			for _, v := range data[i] {
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

// ExtractPercentages returnes the corresponding map with heuristic percentages for each element in the map (in each block)
func (data HeuristicGraph) ExtractPercentages(heuristicsList heuristics.Mask, from, to int32) (perc map[int32][]float64) {
	perc = make(map[int32][]float64, to-from+1)
	list := heuristicsList.ToList()
	for i := from; i <= to; i++ {
		perc[i] = make([]float64, len(list))
		for h, heuristic := range list {
			counter := 0
			for _, v := range data[i] {
				if change, ok := v[heuristic]; ok {
					if change == 0 {
						counter++
					}
				}
				perc[i][h] = float64(counter) / float64(len(data[i]))
			}
		}
	}
	return
}

// ExtractGlobalPercentages returnes the corresponding map with global heuristic percentages for each heuristic
func (data HeuristicGraph) ExtractGlobalPercentages(heuristicsList heuristics.Mask, from, to int32) (perc []float64) {
	list := heuristicsList.ToList()
	perc = make([]float64, len(list))
	for h, heuristic := range list {
		counter, tot := 0, 0
		for i := from; i <= to; i++ {
			for _, v := range data[i] {
				if change, ok := v[heuristic]; ok {
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
