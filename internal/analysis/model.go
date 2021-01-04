package analysis

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model analysis struct
type Model struct {
	gorm.Model
	ID     uuid.UUID `json:"id" gorm:"primarykey;index;unique"`
	UserID uuid.UUID `json:"user_id" gorm:"index;not null"`
	Type   string    `json:"type" gorm:"index"`
} //@name Analysis

// TableName defines default table name
func (m *Model) TableName() string {
	return "analysis"
}
