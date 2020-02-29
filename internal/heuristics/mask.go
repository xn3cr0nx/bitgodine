package heuristics

import "math"

// Mask struct to represent heuristics vulnerability mask
type Mask byte

// VulnerableMask uses bitwise AND operation to apply a mask to vulnerabilities byte to extract value bit by bit
// and returnes true if the vuln byte is vulnerable to passed heuristic
func (v Mask) VulnerableMask(h Heuristic) bool {
	return v&Mask(math.Pow(2, float64(h))) > 0
}

// MergeMasks uses bitwise OR operation to apply a mask to vulnerabilities byte to merge a new mask with updated heuristics
// bit and return the merge between original byte with updated bits
func MergeMasks(source Mask, update Mask) Mask {
	return source | update
}

// Bytes returnes bytes slice enconded mask
func (v Mask) Bytes() []byte {
	return []byte{byte(v)}
}

// MaskFromBytes returnes mask from slice of bytes
func MaskFromBytes(v []byte) Mask {
	return Mask(v[0])
}

// MaskFromPower returnes the mask byte from the power of 2
func MaskFromPower(h Heuristic) Mask {
	return Mask(math.Pow(2, float64(h)))
}

// Sum returnes the updated base mask
func (v *Mask) Sum(m Mask) Mask {
	return (*v) + m
}

// ToList return a list of heuristic integers corresponding to vulnerability byte passed
func (v Mask) ToList() (heuristics []Heuristic) {
	for i := Heuristic(0); i < 8; i++ {
		if v.VulnerableMask(Heuristic(i)) {
			heuristics = append(heuristics, i)
		}
	}
	return
}

// ToHeuristicsList return a list of heuristic names corresponding to vulnerability byte passed
func (v Mask) ToHeuristicsList() (heuristics []string) {
	for i := Heuristic(0); i < 8; i++ {
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
