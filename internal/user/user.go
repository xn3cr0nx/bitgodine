package user

import (
	"time"

	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
)

// Service interface exports available methods for user service
type Service interface {
	GetUser(ID string) (user *Model, err error)
	GetUserByEmail(email string) (user *Model, err error)
	CreateUser(user *Model) (err error)
	NewLogin(ID string) (time.Time, error)
	NewAPIKey(ID, key string) (err error)
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

// GetUser retrieves the user
func (s *service) GetUser(ID string) (*Model, error) {
	var user Model
	if err := s.Repository.Preload("Preferences").Where("ID = ?", ID).Find(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves the user by email
func (s *service) GetUserByEmail(email string) (*Model, error) {
	var user Model
	if err := s.Repository.Preload("Preferences").Where("email = ?", email).Find(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func (s *service) CreateUser(user *Model) error {
	return s.Repository.Model(&Model{}).Create(user).Error
}

// UpdateUser update user object
func (s *service) UpdateUser(ID string, user *Model) error {
	return s.Repository.Model(&Model{}).Where("id = ?", ID).Save(user).Error
}

// NewLogin updates last_login to a new tiemstamp
func (s *service) NewLogin(ID string) (time.Time, error) {
	t := time.Now()
	err := s.Repository.Model(&Model{}).Where("id = ?", ID).Update("last_login", t).Error
	return t, err
}

func (s *service) NewAPIKey(ID, key string) (err error) {
	user := Model{}
	if err = s.Repository.Model(&Model{}).Where("id = ?", ID).Select("api_keys").Find(&user).Error; err != nil {
		return
	}

	err = s.Repository.Model(&Model{}).Where("id = ?", ID).Update("api_keys", append(user.APIKeys, key)).Error
	return
}
