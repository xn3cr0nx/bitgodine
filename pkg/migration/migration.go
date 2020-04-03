package migration

import (
	"github.com/xn3cr0nx/bitgodine_server/pkg/abuse"
	"github.com/xn3cr0nx/bitgodine_server/pkg/cluster"
	"github.com/xn3cr0nx/bitgodine_server/pkg/postgres"
	"github.com/xn3cr0nx/bitgodine_server/pkg/tag"
)

// Migration sets up initial migration of tags involved tables
func Migration(pg *postgres.Pg) (err error) {
	err = pg.DB.AutoMigrate(&tag.Model{}, &abuse.Model{}, &cluster.Model{}).Error
	return
}
