package heuristics

// Heuristic type define a enum on implemented heuristics
type Heuristic int

const (
	Peeling Heuristic = iota
	PowerOfTen
	OptimalChange
	AddressType
	AddressReuse
	Locktime
	ClientBehaviour
	Forward
	Backward
)

// SetCardinality returnes the cardinality of the heuristics set
func SetCardinality() int {
	return int(Backward)
}

func (h Heuristic) String() string {
	heuristics := [...]string{
		"Peeling Chain",
		"Power of Ten",
		"Optimal Change",
		"Address Type",
		"Address Reuse",
		"Locktime",
		"Client Behaviour",
		"Forward",
		"Backward",
	}

	return heuristics[h]
}
