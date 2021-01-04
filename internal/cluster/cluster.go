package cluster

import (
	"fmt"
	"os"

	"github.com/fatih/structs"
	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
)

// Service interface exports available methods for block service
type Service interface {
	GetClusters(output bool) (tags []Model, err error)
	CreateCluster(t *Model) (err error)
	GetCluster(address string, output bool) (tags []Model, err error)
}

type service struct {
	Repository *postgres.Pg
	Cache      *cache.Cache
}

// NewService instantiates a new Service layer for customer
func NewService(r *postgres.Pg, c *cache.Cache) *service {
	return &service{
		Repository: r,
		Cache:      c,
	}
}

func printClustersTable(clusters []Model) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(structs.Names(clusters[0]))
	table.SetBorder(false)
	for _, cluster := range clusters {
		table.Append([]string{cluster.Address, fmt.Sprintf("%d", cluster.Cluster)})
	}
	table.Render()
}

// GetClusters retrieve whole clusters list
func (s *service) GetClusters(output bool) (clusters []Model, err error) {
	if err = s.Repository.Find(&clusters).Error; err != nil {
		return
	}

	if output {
		printClustersTable(clusters)
	}

	return
}

// CreateCluster creates a new cluster record
func (s *service) CreateCluster(t *Model) (err error) {
	err = s.Repository.Model(&Model{}).Create(t).Error
	return
}

// GetCluster retrieve clusters related to passed address
func (s *service) GetCluster(address string, output bool) (clusters []Model, err error) {
	if err = s.Repository.Where("address = ?", address).Find(&clusters).Error; err != nil {
		return
	}

	if output {
		printClustersTable(clusters)
	}

	return
}
