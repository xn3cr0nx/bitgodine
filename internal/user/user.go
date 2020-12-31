package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
)

// Service interface exports available methods for user service
type Service interface {
	GetUserByEmail(email string) (user *Model, err error)
	CreateUser(user *Model) (err error)
}

type service struct {
	Repository *postgres.Pg
}

// NewService instantiates a new Service layer for customer
func NewService(r *postgres.Pg) *service {
	return &service{
		Repository: r,
	}
}

// GetUserByEmail retrieves the user by email
func (s *service) GetUserByEmail(email string) (*Model, error) {
	var user Model
	if err := s.Repository.Preload("preferences").Where("email = ?", email).Find(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func (s *service) CreateUser(user *Model) error {
	return s.Repository.Model(&Model{}).Create(user).Error
}

// NewLogin updates last_login to a new tiemstamp
func (s *service) NewLogin(ID uuid.UUID) (time.Time, error) {
	t := time.Now()
	err := s.Repository.Model(&Model{}).Where("id = ?", ID).Update("last_login", t).Error
	return t, err
}
