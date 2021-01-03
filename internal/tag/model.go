package tag

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// Model of tag struct with validation
type Model struct {
	gorm.Model
	ID       uuid.UUID `json:"id" gorm:"primarykey;index;unique"`
	Address  string    `json:"address" validate:"required,btc_addr|btc_addr_bech32" gorm:"size:64;index;not null"`
	Message  string    `json:"message" validate:"required" gorm:"not null"`
	Nickname string    `json:"nickname,omitempty" validate:"" gorm:"index;not null"`
	Type     string    `json:"type,omitempty" validate:"" gorm:"index;not null"`
	Link     string    `json:"link,omitempty" validate:""`
	Verified bool      `json:"verified,omitempty" validate:"" gorm:"default:false"`
} //@name Tag

// TableName defines default table name
func (m Model) TableName() string {
	return "tags"
}
