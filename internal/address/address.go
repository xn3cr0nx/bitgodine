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

// Service interface exports available methods for tx service
type Service interface {
	GetOccurences(address string) (occurences []string, err error)
	GetFirstOccurenceHeight(address string) (height int32, err error)
}

type service struct {
	Kv    kv.DB
	Cache *cache.Cache
}

// NewService instantiates a new Service layer for customer
func NewService(k kv.DB, c *cache.Cache) *service {
	return &service{
		Kv:    k,
		Cache: c,
	}
}

// IsBitcoinAddress returns true is the string is a bitcoin address
func IsBitcoinAddress(text string) bool {
	re := regexp.MustCompile("^(bc1|[13])[a-zA-HJ-NP-Z0-9]{25,39}$")
	return re.MatchString(text)
}

// GetOccurences returnes an array containing the transactions where the address appears in the blockchain
func (s *service) GetOccurences(address string) (occurences []string, err error) {
	occurences, err = s.Kv.ReadKeysWithPrefix(address + "_")
	if err != nil {
		return
	}
	for _, o := range occurences {
		o = o[strings.LastIndex(o, "_")+1:]
	}
	return
}

// GetFirstOccurenceHeight returnes the height of the block in which the address appeared for the first time
func (s *service) GetFirstOccurenceHeight(address string) (height int32, err error) {
	if cached, ok := s.Cache.Get(address); ok {
		height = cached.(int32)
		return
	}

	resp, err := s.Kv.ReadFirstValueByPrefix(address + "_")
	if err != nil {
		return
	}
	h, err := strconv.Atoi(string(resp))
	if err != nil {
		return
	}
	height = int32(h)

	if !s.Cache.Set(address, height, 1) {
		logger.Error("Cache", errorx.ErrCache, logger.Params{"address": address})
	}
	return
}
