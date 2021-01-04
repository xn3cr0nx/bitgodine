package abuse

import (
	"os"

	"github.com/fatih/structs"
	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// Service interface exports available methods for block service
type Service interface {
	GetAbuses(output bool) (abuses []Model, err error)
	CreateAbuse(abuse *Model) (err error)
	GetAbuse(address string, output bool) (abuses []Model, err error)
	GetAbusedCluster(address string) (clusters []AbusedCluster, err error)
	GetAbusedClusterSet(address string) (clusters []Model, err error)
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

func printAbusesTable(abuses []Model) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(structs.Names(abuses[0]))
	table.SetBorder(false)

	for _, abuse := range abuses {
		table.Append([]string{abuse.Address, abuse.AbuseTypeID, abuse.AbuseTypeOther, abuse.Abuser, abuse.Description, abuse.FromCountry, abuse.FromCountryCode})
	}
	table.Render()
}

// GetAbuses retrieve whole abuses list
func (s *service) GetAbuses(output bool) (abuses []Model, err error) {
	if err = s.Repository.Find(&abuses).Error; err != nil {
		return
	}

	if output {
		printAbusesTable(abuses)
	}

	return
}

// CreateAbuse creates a new abuse record
func (s *service) CreateAbuse(t *Model) (err error) {
	err = s.Repository.Model(&Model{}).Create(t).Error
	return
}

// GetAbuse retrieve abuses related to passed address
func (s *service) GetAbuse(address string, output bool) (abuses []Model, err error) {
	if err = s.Repository.Where("address = ?", address).Find(&abuses).Error; err != nil {
		return
	}

	if output {
		printAbusesTable(abuses)
	}

	return
}

type AbusedCluster struct {
	Model
	Cluster uint64 `json:"cluster" validate:"" gorm:"not null"`
}

// GetAbusedCluster retrieves crossed data between clusters and abuses
func (s *service) GetAbusedCluster(address string) (clusters []AbusedCluster, err error) {
	err = s.Repository.Raw(`SELECT *, c.cluster 
		FROM abuses t 
		RIGHT JOIN clusters c 
		ON t.address = c.address 
		WHERE t.address = ?`, address).Scan(&clusters).Error
	return
}

// GetAbusedClusterSet retrieves crossed data between clusters and abuses
func (s *service) GetAbusedClusterSet(address string) (clusters []Model, err error) {
	if cached, ok := s.Cache.Get("ca_" + address); ok {
		clusters = cached.([]Model)
		return
	}

	err = s.Repository.Raw(`SELECT abuses.abuser FROM "abuses" 
		LEFT JOIN "clusters" 
		ON abuses.address=clusters.address 
		WHERE clusters.cluster=(
		SELECT cluster FROM clusters WHERE address = ? LIMIT 1
	) GROUP BY abuses.abuser`, address).Scan(&clusters).Error

	if !s.Cache.Set("ca_"+address, clusters, 1) {
		logger.Error("Cache", errorx.ErrCache, logger.Params{"address": address})
	}
	return
}
