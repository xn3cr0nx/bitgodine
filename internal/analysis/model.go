package analysis

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model analysis struct
type Model struct {
	gorm.Model
	ID     uuid.UUID `json:"id" gorm:"primary_key;index;unique"`
	UserID uuid.UUID `json:"user_id" gorm:"index"`
	Type   string    `json:"type" gorm:"index"`
}
