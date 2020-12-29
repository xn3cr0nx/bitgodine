package address

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// IsBitcoinAddress returns true is the string is a bitcoin address
func IsBitcoinAddress(text string) bool {
	re := regexp.MustCompile("^(bc1|[13])[a-zA-HJ-NP-Z0-9]{25,39}$")
	return re.MatchString(text)
}

// GetOccurences returnes an array containing the transactions where the address appears in the blockchain
func GetOccurences(db kv.DB, c *cache.Cache, address string) (occurences []string, err error) {
	occurences, err = db.ReadKeysWithPrefix(address + "_")
	if err != nil {
		return
	}
	for _, o := range occurences {
		o = o[strings.LastIndex(o, "_")+1:]
	}
	return
}

// GetFirstOccurenceHeight returnes the height of the block in which the address appeared for the first time
func GetFirstOccurenceHeight(db kv.DB, c *cache.Cache, address string) (height int32, err error) {
	if cached, ok := c.Get(address); ok {
		height = cached.(int32)
		return
	}

	resp, err := db.ReadFirstValueByPrefix(address + "_")
	if err != nil {
		return
	}
	h, err := strconv.Atoi(string(resp))
	if err != nil {
		return
	}
	height = int32(h)

	if !c.Set(address, height, 1) {
		logger.Error("Cache", errorx.ErrCache, logger.Params{"address": address})
	}
	return
}
