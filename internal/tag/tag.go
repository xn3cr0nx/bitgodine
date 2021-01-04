package tag

import (
	"os"

	"github.com/fatih/color"
	"github.com/fatih/structs"
	"github.com/olekukonko/tablewriter"

	// "github.com/xn3cr0nx/bitgodine/internal/cluster"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// Service interface exports available methods for block service
type Service interface {
	GetTags(output bool) (tags []Model, err error)
	CreateTag(t *Model) (err error)
	GetTag(address string, output bool) (tags []Model, err error)
	GetTaggedCluster(address string) (clusters []TaggedCluster, err error)
	GetTaggedClusterSet(address string) (clusters []Model, err error)
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

func printTagsTable(tags []Model) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(structs.Names(tags[0]))
	table.SetBorder(false)

	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)
	for _, tag := range tags {
		if tag.Verified == true {
			table.Append([]string{tag.Address, tag.Message, tag.Nickname, tag.Type, tag.Link, green.Sprint("✓")})
		} else {
			table.Append([]string{tag.Address, tag.Message, tag.Nickname, tag.Type, tag.Link, red.Sprint("x")})
		}
	}
	table.Render()
}

// GetTags retrieve whole tags list
func (s *service) GetTags(output bool) (tags []Model, err error) {
	if err = s.Repository.Find(&tags).Error; err != nil {
		return
	}

	if output {
		printTagsTable(tags)
	}

	return
}

// CreateTag creates a new tag record
func (s *service) CreateTag(t *Model) (err error) {
	err = s.Repository.Model(&Model{}).Create(t).Error
	return
}

// GetTag retrieve tags related to passed address
func (s *service) GetTag(address string, output bool) (tags []Model, err error) {
	if err = s.Repository.Where("address = ?", address).Find(&tags).Error; err != nil {
		return
	}

	if output {
		printTagsTable(tags)
	}

	return
}

type TaggedCluster struct {
	Model
	Cluster uint64 `json:"cluster" validate:"" gorm:"not null"`
}

// GetTaggedCluster retrieves crossed data between clusters and tags
func (s *service) GetTaggedCluster(address string) (clusters []TaggedCluster, err error) {
	err = s.Repository.Raw(`SELECT *, c.cluster 
		FROM tags t 
		RIGHT JOIN clusters c 
		ON t.address = c.address 
		WHERE t.address = ?`, address).Scan(&clusters).Error
	return
}

// GetTaggedClusterSet retrieves crossed data between clusters and tags
func (s *service) GetTaggedClusterSet(address string) (clusters []Model, err error) {
	if cached, ok := s.Cache.Get("ct_" + address); ok {
		clusters = cached.([]Model)
		return
	}

	err = s.Repository.Raw(`SELECT tags.message, tags.type FROM "tags" 
		LEFT JOIN "clusters" 
		ON tags.address=clusters.address 
		WHERE clusters.cluster=(
			SELECT cluster FROM clusters WHERE address = ? LIMIT 1
		) GROUP BY tags.message, tags.type`, address).Scan(&clusters).Error

	if !s.Cache.Set("ct_"+address, clusters, 1) {
		logger.Error("Cache", errorx.ErrCache, logger.Params{"address": address})
	}
	return
}

// insert into tags (address, message, nickname, type) values ('18VaKMJciWuk61MjPraRouiAqvoPQmrCmc', 'test1', 'binance', 1);
// insert into tags (address, message, nickname, type) values ('1CYxSkLRUqe3cpVDLa8u9UKdetyfEM5gby', 'test2', 'bitfinxe', 1);
// insert into tags (address, message, nickname, type) values ('1vXfhQpD7adQuNePT3k3pnRKFjP58EdpC', 'test3', 'okex', 1);
// insert into tags (address, message, nickname, type) values ('1FeexV6bAHb8ybZjqQMjJrcCrHGW9sb6uF', 'test4', 'jeez', 1);
// insert into clusters (address, cluster) values ('1FeexV6bAHb8ybZjqQMjJrcCrHGW9sb6uF', 250538);
