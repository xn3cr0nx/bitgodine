package user

import (
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
	if err := s.Repository.Where("email = ?", email).Find(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func (s *service) CreateUser(user *Model) error {
	return s.Repository.Model(&Model{}).Create(user).Error
}
