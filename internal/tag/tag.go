package tag

import (
	"os"

	"github.com/fatih/color"
	"github.com/fatih/structs"
	"github.com/labstack/echo/v4"
	"github.com/olekukonko/tablewriter"
	"github.com/xn3cr0nx/bitgodine_server/internal/cluster"
	"github.com/xn3cr0nx/bitgodine_server/pkg/postgres"
)

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
func GetTags(c *echo.Context, output bool) (tags []Model, err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	if err = pg.DB.Find(&tags).Error; err != nil {
		return
	}

	if output {
		printTagsTable(tags)
	}

	return
}

// CreateTag creates a new tag record
func CreateTag(c *echo.Context, t *Model) (err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	err = (*pg).DB.Model(&Model{}).Create(t).Error
	return
}

// GetTag retrieve tags related to passed address
func GetTag(c *echo.Context, address string, output bool) (tags []Model, err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	if err = pg.DB.Where("address = ?", address).Find(&tags).Error; err != nil {
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
func GetTaggedCluster(c *echo.Context, address string) (clusters []TaggedCluster, err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	err = pg.DB.Raw(`SELECT *, c.cluster 
		FROM tags t 
		RIGHT JOIN clusters c 
		ON t.address = c.address 
		WHERE t.address = ?`, address).Find(&clusters).Error
	return
}

// GetTaggedClusterSet retrieves crossed data between clusters and tags
func GetTaggedClusterSet(c *echo.Context, address string) (clusters []TaggedCluster, err error) {
	pg := (*c).Get("pg").(*postgres.Pg)
	cluster, err := cluster.GetCluster(c, address, false)
	if err != nil {
		return
	}
	if len(cluster) == 0 {
		err = echo.NewHTTPError(404, "cluster not found")
		return
	}
	err = pg.DB.Raw(`SELECT c.cluster, c.address, t.message, t.nickname, t.type, t.link, t.verified
		FROM tags t 
		RIGHT JOIN clusters c 
		ON t.address = c.address 
		WHERE c.cluster = ?`, cluster[0].Cluster).Find(&clusters).Error
	return
}
