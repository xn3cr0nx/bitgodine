package abuse

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// Model of abuse struct with validation
type Model struct {
	gorm.Model
	ID              uuid.UUID `json:"id" gorm:"primary_key;index;unique"`
	Address         string    `json:"address" validate:"required,btc_addr|btc_addr_bech32" gorm:"size:64;index;not null"`
	AbuseTypeID     string    `json:"abuse_type_id"`
	AbuseTypeOther  string    `json:"abuse_type_other"`
	Abuser          string    `gorm:"index"`
	Description     string    `json:"description"`
	FromCountry     string    `json:"from_country"`
	FromCountryCode string    `json:"from_country_code"`
}

// TableName defines default table name
func (m Model) TableName() string {
	return "abuses"
}
