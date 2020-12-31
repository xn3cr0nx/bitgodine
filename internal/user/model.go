package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/xn3cr0nx/bitgodine/internal/password"
)

// Model cluster struct with validation
type Model struct {
	gorm.Model
	ID        uuid.UUID `json:"id" gorm:"primary_key;index;unique"`
	Email     string    `json:"email" validate:"required,email" gorm:"index"`
	Username  string    `json:"username" validate:"required" gorm:"index"`
	Password  string    `json:"password" validate:"required,min=10"`
	FirstName string    `json:"first_name" validate:"required,min=2"`
	LastName  string    `json:"last_name" validate:"required,min=2"`

	LastLogin time.Time `json:"last_login" validate:""`

	Verified  bool           `json:"verified,omitempty" gorm:"default:false"`
	Lang      string         `json:"lang,omitempty" gorm:"default:'en'"`
	IsBlocked bool           `json:"isBlocked" gorm:"default:false"`
	APIKeys   pq.StringArray `json:"apiKeys,omitempty" gorm:"type:varchar(255)[]"`
}

// BeforeCreate encrypt the password before creating
func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {
	hash, err := password.Hash(m.Password)
	if err != nil {
		return
	}
	m.Password = hash
	return
}

// TableName defines default table name
func (m *Model) TableName() string {
	return "users"
}