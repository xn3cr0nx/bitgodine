package addresses

import (
	"regexp"

	"github.com/btcsuite/btcutil"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
)

// FirstAppearence returnes true if the passed address appears for the first time in blockchain at the given block height
func FirstAppearence(address *btcutil.Address) (int32, error) {
	height, err := dgraph.GetAddressFirstOccurenceHeight(address)
	if err != nil {
		return 0, err
	}
	return height, nil
}

// IsBitcoinAddress returnes true is the string is a bitcoin address
func IsBitcoinAddress(text string) bool {
	re := regexp.MustCompile("^(bc1|[13])[a-zA-HJ-NP-Z0-9]{25,39}$")
	return re.MatchString(text)
}
