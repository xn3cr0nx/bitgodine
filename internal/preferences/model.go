package preferences

import "github.com/google/uuid"

// Model preferences struct with validation
type Model struct {
	UserID  uuid.UUID `json:"user_id" gorm:"primarykey;index;unique"`
	Theme   string    `json:"theme" gorm:"index;default:blue"`
	Gravity bool      `json:"gravity" gorm:"default:true"`
} //@name Preference

// TableName defines default table name
func (m *Model) TableName() string {
	return "preferences"
}
