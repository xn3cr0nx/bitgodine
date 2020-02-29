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
