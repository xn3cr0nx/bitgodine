package bitcoin

import (
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"

	"github.com/xn3cr0nx/bitgodine/pkg/disjoint"
)

// Clusterizer defines the objects involved in the generation of clusters
type Clusterizer struct {
	clusters  disjoint.DisjointSet
	db        kv.DB
	pg        *postgres.Pg
	cache     *cache.Cache
	interrupt chan int
	done      chan int
}

// NewClusterizer return a new instance to Bitcoin blockchain clusterizer
func NewClusterizer(d disjoint.DisjointSet, db kv.DB, pg *postgres.Pg, c *cache.Cache, interrupt chan int, done chan int) *Clusterizer {
	return &Clusterizer{
		clusters:  d,
		db:        db,
		pg:        pg,
		cache:     c,
		interrupt: interrupt,
		done:      done,
	}
}
