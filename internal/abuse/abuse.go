package abuse

import (
	"os"

	"github.com/fatih/structs"
	"github.com/labstack/echo/v4"
	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine_server/internal/cluster"
	"github.com/xn3cr0nx/bitgodine_server/pkg/postgres"
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
func GetAbusedClusterSet(c *echo.Context, address string) (clusters []AbusedCluster, err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	cluster, err := cluster.GetCluster(c, address, false)
	if err != nil {
		return
	}
	if len(cluster) == 0 {
		err = echo.NewHTTPError(404, "cluster not found")
		return
	}
	err = pg.DB.Raw(`SELECT t.id, t.created_at, t.updated_at, t.deleted_at, c.cluster, c.address, t.message, t.nickname, t.type, t.link, t.verified
		FROM abuses t 
		RIGHT JOIN clusters c 
		ON t.address = c.address 
		WHERE c.cluster = ?`, cluster[0].Cluster).Scan(&clusters).Error
	return
}
