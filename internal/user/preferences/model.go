package preferences

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model preferences struct with validation
type Model struct {
	gorm.Model
	UserID  uuid.UUID `json:"user_id" gorm:"primarykey;index;unique;not null"`
	Theme   string    `json:"theme" gorm:"index;default:blue"`
	Gravity bool      `json:"gravity" gorm:"default:true"`
} //@name Preference

// TableName defines default table name
func (m *Model) TableName() string {
	return "preferences"
}
