package heuristics

import (
	"math"
)

// Mask struct to represent heuristics vulnerability mask
type Mask [3]byte

// VulnerableMask uses bitwise AND operation to apply a mask to vulnerabilities byte to extract value bit by bit
// and returnes true if the vuln byte is vulnerable to passed heuristic
func (v Mask) VulnerableMask(h Heuristic) (vuln bool) {
	if h <= 7 {
		vuln = v[0]&byte(math.Pow(2, float64(h))) > 0
	} else if h <= 15 {
		vuln = v[1]&byte(math.Pow(2, float64(h-8))) > 0
	} else {
		vuln = v[2]&byte(math.Pow(2, float64(h-16))) > 0
	}
	return
}

// MergeMasks uses bitwise OR operation to apply a mask to vulnerabilities byte to merge a new mask with updated heuristics
// bit and return the merge between original byte with updated bits
func MergeMasks(source Mask, update Mask) Mask {
	return [3]byte{byte(source[0] | update[0]), byte(source[1] | update[1]), byte(source[2] | update[2])}
}

// MaskFromPower returnes the mask byte from the power of 2
func MaskFromPower(h Heuristic) (m Mask) {
	if h <= 7 {
		m = [3]byte{byte(math.Pow(2, float64(h))), byte(0), byte(0)}
	} else if h <= 15 {
		m = [3]byte{byte(0), byte(math.Pow(2, float64(h-8))), byte(0)}
	} else {
		m = [3]byte{byte(0), byte(0), byte(math.Pow(2, float64(h-16)))}
	}
	return
}

// Sum returnes the updated base mask
func (v *Mask) Sum(m Mask) Mask {
	(*v) = [3]byte{byte((*v)[0] + m[0]), byte((*v)[1] + m[1]), byte((*v)[2] + m[2])}
	return *v
}

// ToList return a list of heuristic integers corresponding to vulnerability byte passed
func (v Mask) ToList() (heuristics []Heuristic) {
	for i := Heuristic(0); i < SetCardinality(); i++ {
		if v.VulnerableMask(Heuristic(i)) {
			heuristics = append(heuristics, i)
		}
	}
	return
}

// ToHeuristicsList return a list of heuristic names corresponding to vulnerability byte passed
func (v Mask) ToHeuristicsList() (heuristics []string) {
	for i := Heuristic(0); i < SetCardinality(); i++ {
		if v.VulnerableMask(Heuristic(i)) {
			heuristics = append(heuristics, i.String())
		}
	}
	return
}

// FromListToMask convert list of heuristics to mask type
func FromListToMask(list []Heuristic) (m Mask) {
	for _, h := range list {
		m = m.Sum(MaskFromPower(h))
	}
	return
}

// IsCoinbase checks if corresponding condition bit is true
func (v Mask) IsCoinbase() bool {
	return v.VulnerableMask(Coinbase)
}

// IsSelfTransfer checks if corresponding condition bit is true
func (v Mask) IsSelfTransfer() bool {
	return v.VulnerableMask(SelfTransfer)
}

// IsOffByOneBug checks if corresponding condition bit is true
func (v Mask) IsOffByOneBug() bool {
	return v.VulnerableMask(OffByOne)
}

// IsPeelingLike checks if corresponding condition bit is true
func (v Mask) IsPeelingLike() bool {
	return v.VulnerableMask(PeelingLike)
}
