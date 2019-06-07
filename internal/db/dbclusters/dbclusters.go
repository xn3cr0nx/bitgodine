package dbclusters

import "github.com/dgraph-io/badger"

type Config struct {
	Dir string
}

// DbClusters instance of key value store designed to treat clusters of addresses
type DbClusters struct {
	*badger.DB
}

// NewDbClusters creates a new instance of DbClusters
func NewDbClusters(conf *Config) (*DbClusters, error) {
	opts := badger.DefaultOptions
	opts.Dir, opts.ValueDir = conf.Dir, conf.Dir
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &DbClusters{db}, nil
}
