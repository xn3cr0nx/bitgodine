package heuristics

// Map analysis output map for heuristics change output
type Map map[Heuristic]uint32

// MergeMaps maps returnes the merge of two maps
func MergeMaps(a, b Map) (new Map) {
	new = a
	for h, c := range b {
		a[h] = c
	}
	return
}

// MapFromHeuristics intialized a null map with passed heuristics as keys
func MapFromHeuristics(args ...Heuristic) (m Map) {
	m = make(map[Heuristic]uint32)
	for _, h := range args {
		m[h] = 1
	}
	return
}

// ToList return a list of heuristic integers corresponding to vulnerability byte passed
func (v Map) ToList() (heuristics []Heuristic) {
	for i := Heuristic(0); i < SetCardinality(); i++ {
		if _, ok := v[i]; ok {
			heuristics = append(heuristics, i)
		}
	}
	return
}

// ToHeuristicsList return a list of heuristic names corresponding to vulnerability byte passed
func (v Map) ToHeuristicsList() (heuristics []string) {
	for i := Heuristic(0); i < SetCardinality(); i++ {
		if _, ok := v[i]; ok {
			heuristics = append(heuristics, i.String())
		}
	}
	return
}

// IsCoinbase checks if corresponding condition bit is true
func (v Map) IsCoinbase() bool {
	return v[Coinbase] == 1
}

// IsSelfTransfer checks if corresponding condition bit is true
func (v Map) IsSelfTransfer() bool {
	return v[SelfTransfer] == 1
}

// IsOffByOneBug checks if corresponding condition bit is true
func (v Map) IsOffByOneBug() bool {
	return v[OffByOne] == 1
}

// IsPeelingLike checks if corresponding condition bit is true
func (v Map) IsPeelingLike() bool {
	return v[PeelingLike] == 1
}
