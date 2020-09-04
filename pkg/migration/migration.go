package migration

import (
	"github.com/xn3cr0nx/bitgodine/internal/abuse"
	"github.com/xn3cr0nx/bitgodine/internal/cluster"
	"github.com/xn3cr0nx/bitgodine/internal/tag"
	"github.com/xn3cr0nx/bitgodine/pkg/postgres"
)

// Migration sets up initial migration of tags involved tables
func Migration(pg *postgres.Pg) (err error) {
	if !pg.DB.Migrator().HasTable("tags") {
		err = pg.DB.Migrator().CreateTable(&tag.Model{})
	}
	if !pg.DB.Migrator().HasTable("abuses") {
		err = pg.DB.Migrator().CreateTable(&abuse.Model{})
	}
	if !pg.DB.Migrator().HasTable("clusters") {
		err = pg.DB.Migrator().CreateTable(&cluster.Model{})
	}
	return
}
