package cluster

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/gorm"
)

// Model cluster struct with validation
type Model struct {
	gorm.Model
	Address string `json:"address" validate:"required,btc_addr|btc_addr_bech32" gorm:"index"`
	Cluster uint64 `json:"cluster" validate:"" gorm:"index"`
}

// TableName defines default table name
func (m Model) TableName() string {
	return "clusters"
}
