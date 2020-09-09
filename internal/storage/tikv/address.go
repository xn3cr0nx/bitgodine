package tikv

import (
	"errors"
	"strconv"
	"strings"

	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// GetAddressOccurences returns an array containing the transactions where the address appears in the blockchain
func (db *KV) GetAddressOccurences(address string) (occurences []string, err error) {
	occurences, err = db.ReadKeysWithPrefix(address + "_")
	if err != nil {
		return
	}
	for _, o := range occurences {
		o = o[strings.LastIndex(o, "_")+1:]
	}
	return
}

// GetAddressFirstOccurenceHeight returns the height of the block in which the address appeared for the first time
func (db *KV) GetAddressFirstOccurenceHeight(address string) (height int32, err error) {
	if cached, ok := db.cache.Get(address); ok {
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

	if !db.cache.Set(address, height, 1) {
		logger.Error("Cache", errors.New("error caching"), logger.Params{"address": address})
	}
	return
}
