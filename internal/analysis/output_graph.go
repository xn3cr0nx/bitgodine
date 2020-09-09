package analysis

import (
	"fmt"
	"math"

	"github.com/xn3cr0nx/bitgodine/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine/pkg/badger"
)

// OutputGraph alias for struct describing blockchain graph based on heuristics
type OutputGraph map[int32]map[string]heuristics.Map

// ExtractPercentages returnes the corresponding map with heuristic percentages for each element in the map (in each block)
func (g OutputGraph) ExtractPercentages(heuristicsList heuristics.Mask, from, to int32) (perc map[int32][]float64) {
	perc = make(map[int32][]float64, to-from+1)
	list := heuristicsList.ToList()
	for i := from; i <= to; i++ {
		perc[i] = make([]float64, len(list))
		for h, heuristic := range list {
			counter, tot := 0, 0
			for _, v := range g[i] {
				if !v.IsCoinbase() {
					if _, ok := v[heuristic]; ok {
						counter++
					}
					tot++
				}
				if tot == 0 {
					perc[i][h] = 0
				} else {
					perc[i][h] = float64(counter) / float64(tot)
				}
			}
		}
	}
	return
}

// ExtractOffByOneBug returnes the corresponding map with heuristic percentages for each element in the map (in each block)
func (g OutputGraph) ExtractOffByOneBug(heuristicsList heuristics.Mask, from, to int32) (perc map[int32][]float64) {
	perc = make(map[int32][]float64, to-from+1)
	list := heuristicsList.ToList()
	for i := from; i <= to; i++ {
		perc[i] = make([]float64, len(list))
		for h, heuristic := range list {
			counter, tot := 0, 0
			for _, v := range g[i] {
				if !v.IsCoinbase() && v.IsOffByOneBug() {
					if _, ok := v[heuristic]; ok {
						if counter == 0 {
							counter++
						}
					}
					tot++
				}
				perc[i][h] = float64(counter) / float64(tot)
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
				if !v.IsCoinbase() {
					if _, ok := v[heuristic]; ok {
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

// ExtractGlobalOffByOneBug returnes the corresponding map with global heuristic percentages for each heuristic
func (g OutputGraph) ExtractGlobalOffByOneBug(heuristicsList heuristics.Mask, from, to int32) (perc []float64) {
	list := heuristicsList.ToList()
	perc = make([]float64, len(list))
	for h, heuristic := range list {
		counter, tot := 0, 0
		for i := from; i <= to; i++ {
			for _, v := range g[i] {
				if !v.IsCoinbase() {
					if v.IsOffByOneBug() {
						if change, ok := v[heuristic]; ok {
							if change == 0 {
								counter++
							}
						}
					}
					tot++
				}
			}
		}
		perc[h] = float64(counter) / float64(tot)
	}
	return
}

// ExtractGlobalSecureBasisPerc returnes the corresponding map with global heuristic percentages for each heuristic
func (g OutputGraph) ExtractGlobalSecureBasisPerc(heuristicsList heuristics.Mask, from, to int32) (perc []float64) {
	list := heuristicsList.ToList()
	perc = make([]float64, len(list))
	for h, heuristic := range list {
		counter, tot := 0, 0
		for i := from; i <= to; i++ {

			for _, v := range g[i] {
				var reuse uint32
				isReuse := false
				var shadow uint32
				isShadow := false
				reuse, ok := v[heuristics.Index("Address Reuse")]
				if ok {
					isReuse = true
				}

				shadow, ok = v[heuristics.Index("Shadow")]
				// Check on vout length cause if shadow applies, the reused address is most likely provided
				// to receive a payment, hence the change address is not the already used address, but the other
				// and we can be sure about this condition just when vout length is exactly 2
				if ok && v[heuristics.Index("OffByOne")] == 0 {
					isShadow = true
					if shadow == 0 {
						shadow = 1
					} else {
						shadow = 0
					}
				}

				if !v.IsCoinbase() {
					if isReuse || isShadow {
						if _, ok := v[heuristic]; ok {
							if v[heuristic] == reuse || v[heuristic] == shadow {
								counter++
							}
						}
					}
					tot++
				}
			}
		}
		perc[h] = float64(counter) / float64(tot)
	}
	return
}

func checkOutputs(m heuristics.Map) (r bool) {
	var list []uint32
	for heuristic, change := range m {
		if heuristic > 8 {
			continue
		}
		list = append(list, change)
	}
	if len(list) > 0 {
		first := list[0]
		for _, e := range list {
			if e != first {
				r = true
				break
			}
		}
	}
	return
}

// ExtractCombinationPercentages returnes the corresponding map with global heuristic percentages for each heuristic
func (g OutputGraph) ExtractCombinationPercentages(heuristicsList heuristics.Mask, from, to int32) (perc map[string]float64) {
	list := heuristicsList.ToList()
	perc = make(map[string]float64, int(math.Pow(2, float64(len(list)))))
	prev := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))
	tot := 0
	for i := from; i <= to; i++ {
		for _, v := range g[i] {
			// include only if all change outputs are the same
			if checkOutputs(v) {
				continue
			}

			if !v.IsCoinbase() {
				prev[v.ToMask()[0]] = prev[v.ToMask()[0]] + 1
				tot++
			}
		}
	}
	for k, v := range prev {
		perc[fmt.Sprintf("%b", k)] = v / float64(tot)
	}
	return
}

func extractMajorityMask(m heuristics.Map, basis uint32) (r heuristics.Map) {
	majority := make(heuristics.Map, len(m))
	for k, v := range m {
		majority[k] = v
	}
	for _, n := range []heuristics.Heuristic{5, 6, 8, 9, 10, 11, 17, 18, 19, 20} {
		delete(majority, n)
	}

	clusters := make(map[uint32][]heuristics.Heuristic)
	for heuristic, change := range majority {
		clusters[change] = append(clusters[change], heuristic)
	}
	if len(clusters) == 0 {
		return
	}

	var output uint32
	var max []heuristics.Heuristic
	multiple := true
	for change, cluster := range clusters {
		if len(cluster) > len(max) {
			max = cluster
			output = change
			multiple = false
		} else if len(cluster) == len(max) {
			multiple = true
		}
	}
	if multiple || output != basis {
		return
	}

	r = make(heuristics.Map, len(max))
	for _, heuristic := range max {
		r[heuristic] = output
	}

	return
}

// ExtractGlobalFullMajorityVotingPerc returnes the corresponding map with global heuristic percentages for each heuristic
func (g OutputGraph) ExtractGlobalFullMajorityVotingPerc(heuristicsList heuristics.Mask, from, to int32) (perc map[string]float64) {
	list := heuristicsList.ToList()
	perc = make(map[string]float64, int(math.Pow(2, float64(len(list)))))
	prev := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))
	tot := 0
	for i := from; i <= to; i++ {

		for _, v := range g[i] {
			var reuse uint32
			isReuse := false
			var shadow uint32
			isShadow := false

			reuse, ok := v[heuristics.Index("Address Reuse")]
			if ok {
				isReuse = true
			}
			shadow, ok = v[heuristics.Index("Shadow")]
			// Check on vout length cause if shadow applies, the reused address is most likely provided
			// to receive a payment, hence the change address is not the already used address, but the other
			// and we can be sure about this condition just when vout length is exactly 2
			if ok && v[heuristics.Index("OffByOne")] == 0 {
				isShadow = true
				if shadow == 0 {
					shadow = 1
				} else {
					shadow = 0
				}
			}

			if !v.IsCoinbase() && (isReuse || isShadow) {
				var majority heuristics.Map
				if isReuse {
					majority = extractMajorityMask(v, reuse)
				} else {
					majority = extractMajorityMask(v, shadow)
				}
				prev[majority.ToMask()[0]] = prev[majority.ToMask()[0]] + 1
				tot++
			}
		}
	}
	for k, v := range prev {
		perc[fmt.Sprintf("%b", k)] = v / float64(tot)
	}
	return
}

// ExtractGlobalMajorityVotingPerc returnes the corresponding map with global heuristic percentages for each heuristic
func (g OutputGraph) ExtractGlobalMajorityVotingPerc(heuristicsList heuristics.Mask, from, to int32) (perc map[string]float64) {
	list := heuristicsList.ToList()
	perc = make(map[string]float64, int(math.Pow(2, float64(len(list)))))
	prev := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))
	counters := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))

	for i := from; i <= to; i++ {
		for _, v := range g[i] {
			var reuse uint32
			isReuse := false
			var shadow uint32
			isShadow := false

			reuse, ok := v[heuristics.Index("Address Reuse")]
			if ok {
				isReuse = true
			}
			shadow, ok = v[heuristics.Index("Shadow")]
			// Check on vout length cause if shadow applies, the reused address is most likely provided
			// to receive a payment, hence the change address is not the already used address, but the other
			// and we can be sure about this condition just when vout length is exactly 2
			if ok && v[heuristics.Index("OffByOne")] == 0 {
				isShadow = true
				if shadow == 0 {
					shadow = 1
				} else {
					shadow = 0
				}
			}

			if !v.IsCoinbase() && (isReuse || isShadow) {
				var majority heuristics.Map
				if isReuse {
					majority = extractMajorityMask(v, reuse)
				} else {
					majority = extractMajorityMask(v, shadow)
				}
				prev[majority.ToMask()[0]] = prev[majority.ToMask()[0]] + 1
				counters[v.ToMask()[0]] = counters[v.ToMask()[0]] + 1
			}
		}
	}
	for k, v := range prev {
		perc[fmt.Sprintf("%b", k)] = v / counters[k]
	}
	return
}

func majorityOutput(m heuristics.Map, reducing ...heuristics.Heuristic) (r heuristics.Map, output uint32) {
	majority := make(heuristics.Map, len(m))
	for k, v := range m {
		majority[k] = v
	}

	for _, n := range []heuristics.Heuristic{5, 6, 8, 9, 10, 11, 17, 18, 19, 20} {
		delete(majority, n)
	}

	for _, r := range reducing {
		delete(majority, r)
	}

	clusters := make(map[uint32][]heuristics.Heuristic)
	for heuristic, change := range majority {
		clusters[change] = append(clusters[change], heuristic)
	}
	if len(clusters) == 0 {
		return
	}

	var max []heuristics.Heuristic
	multiple := true
	for change, cluster := range clusters {
		if len(cluster) > len(max) {
			max = cluster
			output = change
			multiple = false
		} else if len(cluster) == len(max) {
			multiple = true
		}
	}
	if multiple {
		return
	}

	r = make(heuristics.Map, len(max))
	for _, heuristic := range max {
		r[heuristic] = output
	}

	// add reduced heurstics to the mask again in order to avoid it to fallback to a reduced group
	// in such a way to keep track of improvements on the majority set
	for _, reduced := range reducing {
		r[reduced] = 1
	}

	return
}

// ExtractGlobalStricMajorityVotingPerc returnes the corresponding map with global heuristic percentages for each heuristic
func (g OutputGraph) ExtractGlobalStricMajorityVotingPerc(heuristicsList heuristics.Mask, from, to int32) (perc map[string]float64) {
	list := heuristicsList.ToList()
	perc = make(map[string]float64, int(math.Pow(2, float64(len(list)))))
	prev := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))
	counters := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))

	for i := from; i <= to; i++ {
		for _, v := range g[i] {
			var reuse uint32
			isReuse := false
			var shadow uint32
			isShadow := false

			reuse, ok := v[heuristics.Index("Address Reuse")]
			if ok {
				isReuse = true
			}
			shadow, ok = v[heuristics.Index("Shadow")]
			// Check on vout length cause if shadow applies, the reused address is most likely provided
			// to receive a payment, hence the change address is not the already used address, but the other
			// and we can be sure about this condition just when vout length is exactly 2
			if ok && v[heuristics.Index("OffByOne")] == 0 {
				isShadow = true
				if shadow == 0 {
					shadow = 1
				} else {
					shadow = 0
				}
			}

			if !v.IsCoinbase() && (isReuse || isShadow) {
				majority, output := majorityOutput(v)
				if output == reuse || output == shadow {
					prev[majority.ToMask()[0]] = prev[majority.ToMask()[0]] + 1
				}
				counters[majority.ToMask()[0]] = counters[majority.ToMask()[0]] + 1
			}
		}
	}
	for k, v := range prev {
		perc[fmt.Sprintf("%b", k)] = v / counters[k]
	}
	return
}

// MajorityFullAnalysis returnes the corresponding map with global heuristic percentages for each heuristic
func (g OutputGraph) MajorityFullAnalysis(heuristicsList heuristics.Mask, from, to int32, reducing ...heuristics.Heuristic) AnalysisSet {
	list := heuristicsList.ToList()
	localPerc := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))
	perc := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))
	prev := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))
	combinations := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))
	counters := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))
	localCounters := make(map[byte]float64, int(math.Pow(2, float64(len(list)))))

	for i := from; i <= to; i++ {
		for _, v := range g[i] {
			var reuse uint32
			isReuse := false
			var shadow uint32
			isShadow := false

			reuse, ok := v[heuristics.Index("Address Reuse")]
			if ok {
				isReuse = true
			}

			shadow, ok = v[heuristics.Index("Shadow")]
			// Check on vout length cause if shadow applies, the reused address is most likely provided
			// to receive a payment, hence the change address is not the already used address, but the other
			// and we can be sure about this condition just when vout length is exactly 2
			if ok && v[heuristics.Index("OffByOne")] == 0 {
				isShadow = true
				if shadow == 0 {
					shadow = 1
				} else {
					shadow = 0
				}
			}

			majority, output := majorityOutput(v, reducing...)
			if !v.IsCoinbase() && (isReuse || isShadow) {
				if output == reuse || output == shadow {
					prev[majority.ToMask()[0]] = prev[majority.ToMask()[0]] + 1
				}
				localCounters[majority.ToMask()[0]] = localCounters[majority.ToMask()[0]] + 1
			}
			counters[majority.ToMask()[0]] = counters[majority.ToMask()[0]] + 1
			combinations[v.ToMask()[0]] = combinations[v.ToMask()[0]] + 1
		}
	}

	for k, v := range counters {
		localPerc[k] = prev[k] / localCounters[k]
		perc[k] = prev[k] / v
	}
	return AnalysisSet{LocalPercentages: localPerc, Percentages: perc, LocalCounters: localCounters, Counters: counters, Combinations: combinations}
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
			newGraph[i] = make(map[string]heuristics.Map, len(a.Vulnerabilites.(OutputGraph)[i]))
			for tx, v := range a.Vulnerabilites.(OutputGraph)[i] {
				newGraph[i][tx] = heuristics.MergeMaps(v, g[i][tx])
			}
		}
	}
	return newGraph
}
