package heuristics

import (
	"math"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/dgraph"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/backward"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/behaviour"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/forward"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/locktime"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/optimal"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/peeling"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/power"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/reuse"
	class "github.com/xn3cr0nx/bitgodine_server/internal/heuristics/type"
)

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

// List returnes the list of heuristics
func List() (heuristics []string) {
	for h := 0; h < SetCardinality(); h++ {
		heuristics = append(heuristics, Heuristic(h).String())
	}
	return
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

// Index returns the index corresponding the heuristic
func Index(r string) int {
	for i := 0; i <= SetCardinality(); i++ {
		if Heuristic(i).String() == r {
			return int(i)
		}
	}
	return -1
}

// VulnerableFunction returnes vulnerable function to be applied to analysis
func VulnerableFunction(h string) func(*dgraph.Dgraph, *models.Tx) bool {
	functions := map[string](func(*dgraph.Dgraph, *models.Tx) bool){
		"Peeling Chain":    peeling.Vulnerable,
		"Power of Ten":     power.Vulnerable,
		"Optimal Change":   optimal.Vulnerable,
		"Address Type":     class.Vulnerable,
		"Address Reuse":    reuse.Vulnerable,
		"Locktime":         locktime.Vulnerable,
		"Client Behaviour": behaviour.Vulnerable,
		"Forward":          forward.Vulnerable,
		"Backward":         backward.Vulnerable,
	}
	return functions[h]
}

// ToHeuristicsList return a list of heuristic names corresponding to vulnerability byte passed
func ToHeuristicsList(v byte) (heuristics []string) {
	for i := 0; i < 8; i++ {
		// bitwise AND operation applies a mask to vulnerabilities byte to extract value bit by bit
		if v&byte(math.Pow(2, float64(i))) > 0 {
			heuristics = append(heuristics, Heuristic(i).String())
		}
	}
	return
}
