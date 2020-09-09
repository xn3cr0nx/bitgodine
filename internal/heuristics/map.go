package heuristics

// Map analysis output map for heuristics change output
type Map map[Heuristic]uint32

// MergeMaps maps returns the merge of two maps
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

// ToMask converts the Map to a Mask ignoring values
func (v Map) ToMask() (m Mask) {
	for h := range v {
		m.Sum(MaskFromPower(h))
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

// MajorityOutput extract the majority output set from map
func (v Map) MajorityOutput(reducing ...Heuristic) (r Map, output uint32) {
	majority := make(Map, len(v))
	for key, value := range v {
		majority[key] = value
	}

	for _, n := range []Heuristic{5, 6, 8, 9, 10, 11, 17, 18, 19, 20} {
		delete(majority, n)
	}

	for _, r := range reducing {
		delete(majority, r)
	}

	clusters := make(map[uint32][]Heuristic)
	for heuristic, change := range majority {
		clusters[change] = append(clusters[change], heuristic)
	}
	if len(clusters) == 0 {
		return
	}

	var max []Heuristic
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

	r = make(Map, len(max))
	for _, heuristic := range max {
		r[heuristic] = output
	}

	// // add reduced heurstics to the mask again in order to avoid it to fallback to a reduced group
	// // in such a way to keep track of improvements on the majority set
	// for _, reduced := range reducing {
	// 	r[reduced] = m[reduced]
	// }

	return
}

// // MajorityLikelihood mapp to define majority voting probability
// var MajorityLikelihood = map[byte]float64{
// 	byte(0b10010111): 49.23,
// 	byte(0b10001100): 53.45,
// 	byte(0b10011):    100.00,
// 	byte(0b10000):    99.31,
// 	byte(0b10000111): 50.41,
// 	byte(0b10101):    99.62,
// 	byte(0b10100):    99.16,
// 	byte(0b10000100): 18.03,
// 	byte(0b10001101): 51.76,
// 	byte(0b10000110): 27.61,
// 	byte(0b10110):    100.00,
// 	byte(0b10001):    97.92,
// 	byte(0b111):      99.29,
// 	byte(0b0):        72.63,
// 	byte(0b10010010): 79.88,
// 	byte(0b10010110): 56.37,
// 	byte(0b10):       96.06,
// 	byte(0b10001001): 52.50,
// 	byte(0b10010101): 52.19,
// 	byte(0b1000):     89.43,
// 	byte(0b11001):    100.00,
// 	byte(0b10010000): 61.63,
// 	byte(0b10111):    100.00,
// 	byte(0b100):      61.35,
// 	byte(0b10011101): 100.00,
// 	byte(0b10011000): 100.00,
// 	byte(0b10000010): 48.04,
// 	byte(0b11101):    100.00,
// 	byte(0b11000):    100.00,
// 	byte(0b10011001): 100.00,
// 	byte(0b10011100): 100.00,
// 	byte(0b1101):     99.37,
// 	byte(0b110):      92.72,
// 	byte(0b10000000): 44.70,
// 	byte(0b10010100): 60.08,
// 	byte(0b1):        52.88,
// 	byte(0b10010011): 48.54,
// 	byte(0b10000011): 49.76,
// 	byte(0b10010001): 53.65,
// 	byte(0b10000001): 50.57,
// 	byte(0b10010):    100.00,
// 	byte(0b11100):    100.00,
// 	byte(0b10001000): 65.27,
// 	byte(0b1100):     82.28,
// 	byte(0b101):      93.27,
// 	byte(0b11):       99.86,
// 	byte(0b10000101): 50.94,
// 	byte(0b1001):     89.21,
// }

// MajorityLikelihood mapp to define majority voting probability
var MajorityLikelihood = map[byte]float64{
	byte(0b10011101): 100.00,
	byte(0b10011100): 100.00,
	byte(0b10011):    96.06,
	byte(0b10111):    93.85,
	byte(0b10010000): 99.99,
	byte(0b10001101): 100.00,
	byte(0b10001):    94.18,
	byte(0b11000):    100.00,
	byte(0b10001001): 100.00,
	byte(0b10010001): 100.00,
	byte(0b10010010): 100.00,
	byte(0b10110):    91.87,
	byte(0b100):      61.50,
	byte(0b1):        58.32,
	byte(0b1001):     83.78,
	byte(0b10011001): 100.00,
	byte(0b1000):     90.33,
	byte(0b10010101): 100.00,
	byte(0b10001000): 99.92,
	byte(0b11):       96.09,
	byte(0b0):        87.23,
	byte(0b10000):    96.57,
	byte(0b10000110): 100.00,
	byte(0b10001100): 99.93,
	byte(0b101):      80.60,
	byte(0b10000001): 100.00,
	byte(0b1101):     80.50,
	byte(0b10010100): 100.00,
	byte(0b1100):     79.09,
	byte(0b10000000): 95.66,
	byte(0b10000010): 100.00,
	byte(0b10000100): 99.76,
	byte(0b10):       96.05,
	byte(0b10000101): 100.00,
	byte(0b10010110): 100.00,
	byte(0b11100):    100.00,
	byte(0b111):      72.94,
	byte(0b11001):    100.00,
	byte(0b10011000): 100.00,
	byte(0b10000111): 100.00,
	byte(0b110):      59.63,
	byte(0b10010111): 100.00,
	byte(0b10100):    78.78,
	byte(0b10000011): 100.00,
	byte(0b11101):    100.00,
	byte(0b10101):    66.49,
	byte(0b10010):    99.02,
	byte(0b10010011): 100.00,
}
