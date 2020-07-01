package abuse

import (
	"errors"
	"os"

	"github.com/fatih/structs"
	"github.com/labstack/echo/v4"
	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/postgres"
)

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
func GetAbuses(c *echo.Context, output bool) (abuses []Model, err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	if err = pg.DB.Find(&abuses).Error; err != nil {
		return
	}

	if output {
		printAbusesTable(abuses)
	}

	return
}

// CreateAbuse creates a new abuse record
func CreateAbuse(c *echo.Context, t *Model) (err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	err = (*pg).DB.Model(&Model{}).Create(t).Error
	return
}

// GetAbuse retrieve abuses related to passed address
func GetAbuse(c *echo.Context, address string, output bool) (abuses []Model, err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	if err = pg.DB.Where("address = ?", address).Find(&abuses).Error; err != nil {
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
func GetAbusedCluster(c *echo.Context, address string) (clusters []AbusedCluster, err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	err = pg.DB.Raw(`SELECT *, c.cluster 
		FROM abuses t 
		RIGHT JOIN clusters c 
		ON t.address = c.address 
		WHERE t.address = ?`, address).Scan(&clusters).Error
	return
}

// GetAbusedClusterSet retrieves crossed data between clusters and abuses
func GetAbusedClusterSet(c *echo.Context, address string) (clusters []Model, err error) {
	ch := (*c).Get("cache").(*cache.Cache)
	if cached, ok := ch.Get("ca_" + address); ok {
		clusters = cached.([]Model)
		return
	}

	pg := (*c).Get("pg").(*postgres.Pg)
	err = pg.DB.Raw(`SELECT abuses.abuser FROM "abuses" 
		LEFT JOIN "clusters" 
		ON abuses.address=clusters.address 
		WHERE clusters.cluster=(
		SELECT cluster FROM clusters WHERE address = ? LIMIT 1
	) GROUP BY abuses.abuser`, address).Scan(&clusters).Error

	if !ch.Set("ca_"+address, clusters, 1) {
		logger.Error("Cache", errors.New("error caching"), logger.Params{"address": address})
	}
	return
}
