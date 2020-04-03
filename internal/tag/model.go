package tag

import (
	"github.com/jinzhu/gorm"
)

// Model of tag struct with validation
type Model struct {
	gorm.Model
	Address  string `json:"address" validate:"required,btc_addr|btc_addr_bech32" gorm:"size:64;index;not null"`
	Message  string `json:"message" validate:"required" gorm:"index;not null"`
	Nickname string `json:"nickname,omitempty" validate:"" gorm:"index;not null"`
	Type     string `json:"type,omitempty" validate:"" gorm:"index;not null"`
	Link     string `json:"link,omitempty" validate:""`
	Verified bool   `json:"verified,omitempty" validate:"" gorm:"default:false"`
}
