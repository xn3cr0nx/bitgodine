package heuristics

import (
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/backward"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/behaviour"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/forward"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/locktime"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/optimal"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/peeling"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/power"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/reuse"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics/shadow"
	class "github.com/xn3cr0nx/bitgodine/internal/heuristics/type"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/models"
)

// Heuristic type define a enum on implemented heuristics
type Heuristic int

const (
	Locktime Heuristic = iota
	Peeling
	PowerOfTen
	OptimalChange
	AddressType
	AddressReuse
	Shadow
	ClientBehaviour

	ExactAmount
	Backward
	Forward
	h12
	h13
	h14
	h15
	h16

	Coinbase
	SelfTransfer
	OffByOne
	PeelingLike
	h21
	h22
	h23
	h24
)

// SetCardinality returns the cardinality of the heuristics set
func SetCardinality() Heuristic {
	// return int(Forward) + 1
	return ClientBehaviour + 1
}

func (h Heuristic) String() string {
	heuristics := [...]string{
		"Locktime",
		"Peeling Chain",
		"Power of Ten",
		"Optimal Change",
		"Address Type",
		"Address Reuse",
		"Shadow",
		"Client Behaviour",
		"Exact Amount",
		"Backward",
		"Forward",
		"",
		"",
		"",
		"",
		"",
		"Coinbase",
		"SelfTransfer",
		"OffByOne",
		"PeelingLike",
		"",
		"",
		"",
		"",
	}
	return heuristics[h]
}

// Abbreviation returns vulnerable function to be applied to analysis
func Abbreviation(a string) Heuristic {
	abbreviations := map[string]Heuristic{
		"locktime": Locktime,
		"peeling":  Peeling,
		"power":    PowerOfTen,
		"optimal":  OptimalChange,
		"type":     AddressType,
		"reuse":    AddressReuse,
		"shadow":   Shadow,
		"client":   ClientBehaviour,
		"exact":    ExactAmount,
		"forward":  Forward,
		"backward": Backward,
	}
	return abbreviations[a]
}

// List returns the list of heuristics
func List() (heuristics []Heuristic) {
	// SetCardinality is used in ToList and ToHeuristicsList in mask
	for h := Heuristic(0); h < SetCardinality(); h++ {
		heuristics = append(heuristics, h)
	}
	return
}

// Index returns the index corresponding the heuristic
func Index(h string) Heuristic {
	functions := map[string](Heuristic){
		"Locktime":         Locktime,
		"Peeling Chain":    Peeling,
		"Power of Ten":     PowerOfTen,
		"Optimal Change":   OptimalChange,
		"Address Type":     AddressType,
		"Address Reuse":    AddressReuse,
		"Shadow":           Shadow,
		"Client Behaviour": ClientBehaviour,
		"Exact Amount":     ExactAmount,
		"Backward":         Backward,
		"Forward":          Forward,
		"h12":              h12,
		"h13":              h13,
		"h14":              h14,
		"h15":              h15,
		"h16":              h16,
		"Coinbase":         Coinbase,
		"SelfTransfer":     SelfTransfer,
		"OffByOne":         OffByOne,
		"PeelingLike":      PeelingLike,
		"h21":              h21,
		"h22":              h22,
		"h23":              h23,
		"h24":              h24,
	}
	return functions[h]
}

// VulnerableFunction returns vulnerable function to be applied to analysis
func (h Heuristic) VulnerableFunction() func(storage.DB, *models.Tx) bool {
	functions := map[Heuristic](func(storage.DB, *models.Tx) bool){
		Locktime:        locktime.Vulnerable,
		Peeling:         peeling.Vulnerable,
		PowerOfTen:      power.Vulnerable,
		OptimalChange:   optimal.Vulnerable,
		AddressType:     class.Vulnerable,
		AddressReuse:    reuse.Vulnerable,
		Shadow:          shadow.Vulnerable,
		ClientBehaviour: behaviour.Vulnerable,
		// "Exact Amount": 		self.Vulnerable,
		Forward:  forward.Vulnerable,
		Backward: backward.Vulnerable,
	}
	return functions[h]
}

// ChangeFunction returns change output function to be applied to analysis
func (h Heuristic) ChangeFunction() func(storage.DB, *models.Tx) ([]uint32, error) {
	functions := map[Heuristic](func(storage.DB, *models.Tx) ([]uint32, error)){
		Locktime:        locktime.ChangeOutput,
		Peeling:         peeling.ChangeOutput,
		PowerOfTen:      power.ChangeOutput,
		OptimalChange:   optimal.ChangeOutput,
		AddressType:     class.ChangeOutput,
		AddressReuse:    reuse.ChangeOutput,
		Shadow:          shadow.ChangeOutput,
		ClientBehaviour: behaviour.ChangeOutput,
		// "Exact Amount": 		self.ChangeOutput,
		Forward:  forward.ChangeOutput,
		Backward: backward.ChangeOutput,
	}
	return functions[h]
}

// ConditionFunction returns change output function to be applied to analysis
func (h Heuristic) ConditionFunction() func(*models.Tx) bool {
	functions := map[Heuristic](func(*models.Tx) bool){
		Coinbase:     coinbaseCondition,
		SelfTransfer: selfTransferCondition,
		OffByOne:     offByOneBugCondition,
		PeelingLike:  peeling.PeelingLikeCondition,
	}
	return functions[h]
}

// Apply applies the heuristic specified to the passed transaction
func (h Heuristic) Apply(db storage.DB, tx models.Tx, vuln *Mask) {
	if h.VulnerableFunction()(db, &tx) {
		vuln.Sum(MaskFromPower(h))
	}
}

// ApplyFullSet applies the set of heuristics to the passed transaction
func ApplyFullSet(db storage.DB, tx models.Tx, vuln *Mask) {
	for _, h := range List() {
		h.Apply(db, tx, vuln)
	}
}

// ApplySet applies the set of passed heuristics to the passed transaction
func ApplySet(db storage.DB, tx models.Tx, heuristicsList Mask, vuln *Mask) {
	for _, h := range heuristicsList.ToList() {
		h.Apply(db, tx, vuln)
	}
}

// ApplyChange applies the heuristic specified to the passed transaction
func (h Heuristic) ApplyChange(db storage.DB, tx models.Tx, vuln *Map) {
	c, err := h.ChangeFunction()(db, &tx)
	if err != nil {
		return
	}
	if len(c) == 1 {
		(*vuln)[h] = c[0]
	}
}

// ApplyChangeSet applies the set of passed heuristics to the passed transaction
func ApplyChangeSet(db storage.DB, tx models.Tx, heuristicsList Mask, vuln *Map) {
	for _, h := range heuristicsList.ToList() {
		h.ApplyChange(db, tx, vuln)
	}
}

// ApplyCondition applies the heuristic specified to the passed transaction
func (h Heuristic) ApplyCondition(db storage.DB, tx models.Tx, vuln *Mask) {
	if h.ConditionFunction()(&tx) {
		vuln.Sum(MaskFromPower(h))
	}
}

// ApplyConditionSet applies the set of passed heuristics to the passed transaction
func ApplyConditionSet(db storage.DB, tx models.Tx, vuln *Mask) {
	for _, h := range conditionsList() {
		h.ApplyCondition(db, tx, vuln)
	}
}

// ApplyChangeCondition applies the heuristic specified to the passed transaction
func (h Heuristic) ApplyChangeCondition(db storage.DB, tx models.Tx, vuln *Map) {
	if h.ConditionFunction()(&tx) {
		(*vuln)[h] = 1
	}
}

// ApplyChangeConditionSet applies the set of passed heuristics to the passed transaction
func ApplyChangeConditionSet(db storage.DB, tx models.Tx, vuln *Map) {
	for _, h := range conditionsList() {
		h.ApplyChangeCondition(db, tx, vuln)
	}
}
