package cluster

import (
	"fmt"
	"os"

	"github.com/fatih/structs"
	"github.com/labstack/echo/v4"
	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine/pkg/postgres"
)

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
func GetClusters(c *echo.Context, output bool) (clusters []Model, err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	if err = pg.DB.Find(&clusters).Error; err != nil {
		return
	}

	if output {
		printClustersTable(clusters)
	}

	return
}

// CreateCluster creates a new cluster record
func CreateCluster(c *echo.Context, t *Model) (err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	err = (*pg).DB.Model(&Model{}).Create(t).Error
	return
}

// GetCluster retrieve clusters related to passed address
func GetCluster(c *echo.Context, address string, output bool) (clusters []Model, err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	if err = pg.DB.Where("address = ?", address).Find(&clusters).Error; err != nil {
		return
	}

	if output {
		printClustersTable(clusters)
	}

	return
}
