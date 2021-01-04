package tx

import (
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/encoding"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// Service interface exports available methods for tx service
type Service interface {
	GetFromHash(hash string) (transaction Tx, err error)
	GetOutputsFromHash(hash string) (outputs []Output, err error)
	GetSpentOutputFromHash(hash string, vout uint32) (output Output, err error)
	GetSpendingFromHash(hash string, vout uint32) (transaction Tx, err error)
	IsSpent(tx string, index uint32) bool
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

// read retrieves tx by hash
func read(db kv.DB, hash string) (transaction Tx, err error) {
	r, err := db.Read(hash)
	if err != nil {
		return
	}
	if err = encoding.Unmarshal(r, &transaction); err != nil {
		return
	}
	return
}

// readFollowing retrieves spending tx of the output based on hash and index
func readFollowing(db kv.DB, hash string, vout uint32) (transaction string, err error) {
	bytes, err := db.Read(hash + "_" + string(vout))
	if err != nil {
		return
	}
	transaction = string(bytes)
	return
}

// GetFromHash return block structure based on block hash
func (s *service) GetFromHash(hash string) (transaction Tx, err error) {
	if cached, ok := s.Cache.Get(hash); ok {
		transaction = cached.(Tx)
		return
	}

	tx, err := read(s.Kv, hash)
	if err != nil {
		return Tx{}, err
	}

	if !s.Cache.Set(transaction.TxID, transaction, 1) {
		logger.Error("Cache", errorx.ErrCache, logger.Params{"hash": transaction.TxID})
	}
	return tx, nil
}

// GetOutputsFromHash retrieves tx's outputs by hash
func (s *service) GetOutputsFromHash(hash string) (outputs []Output, err error) {
	tx, err := s.GetFromHash(hash)
	if err != nil {
		return
	}
	outputs = tx.Vout
	return
}

// GetSpentOutputFromHash retrieves spent tx output based on hash and index
func (s *service) GetSpentOutputFromHash(hash string, vout uint32) (output Output, err error) {
	tx, err := s.GetFromHash(hash)
	if err != nil {
		return
	}
	output = tx.Vout[vout]
	return
}

// GetSpendingFromHash retrieves spending tx of the output based on hash and index
func (s *service) GetSpendingFromHash(hash string, vout uint32) (transaction Tx, err error) {
	spendingHash, err := readFollowing(s.Kv, hash, vout)
	if err != nil {
		return
	}
	transaction, err = s.GetFromHash(spendingHash)
	return
}

// IsSpent returnes true if exists a transaction that takes as input to the new tx
// the output corresponding to the index passed to the function
func (s *service) IsSpent(tx string, index uint32) bool {
	_, err := s.GetSpendingFromHash(tx, index)
	return err == nil
}
