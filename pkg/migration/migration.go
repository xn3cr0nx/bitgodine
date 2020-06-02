package migration

import (
	"github.com/xn3cr0nx/bitgodine_server/internal/abuse"
	"github.com/xn3cr0nx/bitgodine_server/internal/cluster"
	"github.com/xn3cr0nx/bitgodine_server/internal/tag"
	"github.com/xn3cr0nx/bitgodine_server/pkg/postgres"
)

// Migration sets up initial migration of tags involved tables
func Migration(pg *postgres.Pg) (err error) {
	if !pg.DB.HasTable("tags") {
		err = pg.DB.CreateTable(&tag.Model{}).Error
	}
	if !pg.DB.HasTable("abuses") {
		err = pg.DB.CreateTable(&abuse.Model{}).Error
	}
	if !pg.DB.HasTable("clusters") {
		err = pg.DB.CreateTable(&cluster.Model{}).Error
	}
	return
}
