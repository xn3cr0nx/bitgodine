package address

import (
	"regexp"
)

// IsBitcoinAddress returnes true is the string is a bitcoin address
func IsBitcoinAddress(text string) bool {
	re := regexp.MustCompile("^(bc1|[13])[a-zA-HJ-NP-Z0-9]{25,39}$")
	return re.MatchString(text)
}
