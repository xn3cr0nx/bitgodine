package cluster

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/xn3cr0nx/bitgodine_server/internal/cluster"
)

// Model structure to wrap checkbitcoinaddress tags
type Model struct {
	cluster.Model
}
