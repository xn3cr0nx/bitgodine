package migration

import (
	"github.com/xn3cr0nx/bitgodine_server/pkg/abuse"
	"github.com/xn3cr0nx/bitgodine_server/pkg/cluster"
	"github.com/xn3cr0nx/bitgodine_server/pkg/postgres"
	"github.com/xn3cr0nx/bitgodine_server/pkg/tag"
)

// Migration sets up initial migration of tags involved tables
func Migration(pg *postgres.Pg) (err error) {
	if !pg.DB.HasTable("tags") {
		err = pg.DB.Table("tags").CreateTable(&tag.Model{}).Error
	}
	if !pg.DB.HasTable("abuses") {
		err = pg.DB.Table("abuses").CreateTable(&abuse.Model{}).Error
	}
	if !pg.DB.HasTable("clusters") {
		err = pg.DB.Table("clusters").CreateTable(&cluster.Model{}).Error
	}
	return
}
