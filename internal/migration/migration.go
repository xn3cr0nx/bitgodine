package migration

import (
	"github.com/xn3cr0nx/bitgodine/internal/abuse"
	"github.com/xn3cr0nx/bitgodine/internal/analysis"
	"github.com/xn3cr0nx/bitgodine/internal/cluster"
	"github.com/xn3cr0nx/bitgodine/internal/preferences"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/internal/tag"
	"github.com/xn3cr0nx/bitgodine/internal/user"
)

// Migration sets up initial migration of tags involved tables
func Migration(pg *postgres.Pg) (err error) {
	if !pg.DB.Migrator().HasTable("users") {
		err = pg.DB.Migrator().CreateTable(&user.Model{})
	}
	if !pg.DB.Migrator().HasTable("preferences") {
		err = pg.DB.Migrator().CreateTable(&preferences.Model{})
	}
	if !pg.DB.Migrator().HasTable("analysis") {
		err = pg.DB.Migrator().CreateTable(&analysis.Model{})
	}
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
