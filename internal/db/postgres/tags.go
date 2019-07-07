package postgres

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Address defines address record con addresses table
type Address struct {
	Address string `gorm:"primary_key;size:64;index;not null;unique"`
}

// Tag defines tag record for tags table
type Tag struct {
	gorm.Model
	Address   string  `gorm:"size:64;index;not null"`
	Addresses Address `gorm:"foreignkey:Address;association_foreignkey:Address"`
	Tag       string  `gorm:"index;not null"`
	Meta      string  `gorm:"size:255"`
	Verified  bool    `gorm:"default:false"`
}

// Cluster defines cluster record for clusters table
type Cluster struct {
	gorm.Model
	Cluster int    `gorm:"not null"`
	Name    string `gorm:"size:255"`
	Address string `gorm:"primary_key;size:64;index;not null;unique"`
	// Addresses []Address `gorm:"foreignkey:Address;association_foreignkey:Address"`
}

// Migration sets up initial migration of tags involved tables
func (pg *Pg) Migration() error {
	pg.DB.AutoMigrate(&Address{}, &Tag{}, &Cluster{})
	pg.DB.Model(&Tag{}).AddForeignKey("address", "addresses(address)", "RESTRICT", "RESTRICT")
	// pg.DB.Model(&Cluster{}).AddForeignKey("address", "addresses(address)", "RESTRICT", "RESTRICT")
	return nil
}
