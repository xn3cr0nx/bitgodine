package heuristics

import (
	"math"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
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
func VulnerableFunction(h string) func(storage.DB, *models.Tx) bool {
	functions := map[string](func(storage.DB, *models.Tx) bool){
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

// Apply applies the heuristic specified to the passed transaction
func Apply(db storage.DB, tx *models.Tx, h int, vuln *byte) {
	if VulnerableFunction(Heuristic(h).String())(db, tx) {
		(*vuln) += byte(math.Pow(2, float64(h+1)))
	}
}

// ApplySet applies the set of heuristics to the passed transaction
func ApplySet(db storage.DB, tx *models.Tx, vuln *byte) {
	for h := 0; h < SetCardinality(); h++ {
		Apply(db, tx, h, vuln)
	}
}

// ToList return a list of heuristic names corresponding to vulnerability byte passed
func ToList(v byte) (heuristics []string) {
	for i := 0; i < 8; i++ {
		if VulnerableMask(v, i) {
			heuristics = append(heuristics, Heuristic(i).String())
		}
	}
	return
}

// VulnerableMask uses bitwise AND operation to apply a mask to vulnerabilities byte to extract value bit by bit
// and returnes true if the vuln byte is vulnerable to passed heuristic
func VulnerableMask(v byte, h int) bool {
	return v&byte(math.Pow(2, float64(h))) > 0
}
