package heuristics

import (
	"math"

	"github.com/wcharczuk/go-chart/drawing"
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
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics/shadow"
	class "github.com/xn3cr0nx/bitgodine_server/internal/heuristics/type"
)

// Heuristic type define a enum on implemented heuristics
type Heuristic int

const (
	Locktime Heuristic = iota
	Peeling
	PowerOfTen
	OptimalChange
	// ExactAmount
	AddressType
	AddressReuse
	Shadow
	ClientBehaviour
	// Backward
	// Forward
)

// SetCardinality returnes the cardinality of the heuristics set
func SetCardinality() int {
	// return int(Forward) + 1
	// return int(Backward) + 1
	return int(ClientBehaviour) + 1
}

func (h Heuristic) String() string {
	heuristics := [...]string{
		"Locktime",
		"Peeling Chain",
		"Power of Ten",
		"Optimal Change",
		// "Exact Amount",
		"Address Type",
		"Address Reuse",
		"Shadow",
		"Client Behaviour",
		// "Backward",
		// "Forward",
	}
	return heuristics[h]
}

// Abbreviation returnes vulnerable function to be applied to analysis
func Abbreviation(a string) string {
	abbreviations := map[string]string{
		"locktime": "Locktime",
		"peeling":  "Peeling Chain",
		"power":    "Power of Ten",
		"optimal":  "Optimal Change",
		"exact":    "Exact Amount",
		"type":     "Address Type",
		"reuse":    "Address Reuse",
		"shadow":   "Shadow",
		"client":   "Client Behaviour",
		"forward":  "Forward",
		"backward": "Backward",
	}
	return abbreviations[a]
}

// Color returnes color corresponding to each heuristic
func Color(a string) drawing.Color {
	colors := map[string]drawing.Color{
		"Locktime":         drawing.Color{R: 235, G: 255, B: 162},
		"Peeling Chain":    drawing.Color{R: 0, G: 128, B: 128},
		"Power of Ten":     drawing.Color{R: 141, G: 0, B: 0},
		"Optimal Change":   drawing.Color{R: 201, G: 152, B: 0},
		"Exact Amount":     drawing.Color{R: 86, G: 212, B: 101},
		"Address Type":     drawing.Color{R: 64, G: 0, B: 64},
		"Address Reuse":    drawing.Color{R: 0, G: 255, B: 159},
		"Shadow":           drawing.Color{R: 203, G: 12, B: 89},
		"Client Behaviour": drawing.Color{R: 12, G: 67, B: 131},
		"Forward":          drawing.Color{R: 234, G: 124, B: 76},
		"Backward":         drawing.Color{R: 104, G: 52, B: 171},
	}
	return colors[a]
}

// List returnes the list of heuristics
func List() (heuristics []string) {
	for h := 0; h < SetCardinality(); h++ {
		heuristics = append(heuristics, Heuristic(h).String())
	}
	return
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

// HeuristicIndex returns the index corresponding the heuristic
func HeuristicIndex(r string) Heuristic {
	for i := 0; i <= SetCardinality(); i++ {
		if Heuristic(i).String() == r {
			return Heuristic(i)
		}
	}
	return -1
}

// VulnerableFunction returnes vulnerable function to be applied to analysis
func VulnerableFunction(h string) func(storage.DB, *models.Tx) bool {
	functions := map[string](func(storage.DB, *models.Tx) bool){
		"Locktime":       locktime.Vulnerable,
		"Peeling Chain":  peeling.Vulnerable,
		"Power of Ten":   power.Vulnerable,
		"Optimal Change": optimal.Vulnerable,
		// "Exact Amount": 		self.Vulnerable,
		"Address Type":     class.Vulnerable,
		"Address Reuse":    reuse.Vulnerable,
		"Shadow":           shadow.Vulnerable,
		"Client Behaviour": behaviour.Vulnerable,
		"Forward":          forward.Vulnerable,
		"Backward":         backward.Vulnerable,
	}
	return functions[h]
}

// Apply applies the heuristic specified to the passed transaction
func Apply(db storage.DB, tx models.Tx, h string, vuln *byte) {
	if VulnerableFunction(h)(db, &tx) {
		(*vuln) += byte(math.Pow(2, float64(Index(h))))
	}
}

// ApplyFullSet applies the set of heuristics to the passed transaction
func ApplyFullSet(db storage.DB, tx models.Tx, vuln *byte) {
	for h := 0; h < SetCardinality(); h++ {
		Apply(db, tx, Heuristic(h).String(), vuln)
	}
}

// ApplySet applies the set of passed heuristics to the passed transaction
func ApplySet(db storage.DB, tx models.Tx, heuristicsList []string, vuln *byte) {
	for _, heuristic := range heuristicsList {
		Apply(db, tx, heuristic, vuln)
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

// MergeMask uses bitwise OR operation to apply a mask to vulnerabilities byte to merge a new mask with updated heuristics
// bit and return the merge between original byte with updated bits
func MergeMask(source byte, update byte) byte {
	return source | update
}
