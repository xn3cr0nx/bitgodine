package cluster

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Model cluster struct with validation
type Model struct {
	gorm.Model
	Address string `json:"address" validate:"required,btc_addr|btc_addr_bech32" gorm:"index"`
	Cluster uint64 `json:"cluster" validate:""`
}

// TableName defines default table name
func (m Model) TableName() string {
	return "clusters"
}
