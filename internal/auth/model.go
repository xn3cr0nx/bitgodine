package auth

import (
	"time"

	"github.com/lib/pq"
)

// LoginBody encoded email and password authentication
type LoginBody struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

// UserResp includes relevant user fields for login response
type UserResp struct {
	Email     string `json:"email" validate:"required,email" gorm:"index"`
	Username  string `json:"username" validate:"required" gorm:"index"`
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`

	LastLogin time.Time `json:"last_login" validate:""`

	IsActive  bool           `json:"is_active,omitempty" gorm:"default:false"`
	Lang      string         `json:"lang,omitempty" gorm:"default:'en'"`
	IsBlocked bool           `json:"isBlocked" gorm:"default:false"`
	APIKeys   pq.StringArray `json:"apiKeys,omitempty" gorm:"type:varchar(255)[]"`
}

// LoginResp encoded email and password authentication
type LoginResp struct {
	Token string   `json:"token"`
	User  UserResp `json:"user"`
}

// SignupBody encoded user data for registration
type SignupBody struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,password"`
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`
	Username  string `json:"username" validate:"required,min=2"`
}

// SignupResp signup successfull message
type SignupResp struct {
	Message string `json:"message"`
}

// ChangePasswordBody body request to change password
type ChangePasswordBody struct {
	NewPassword string `json:"new_password" validate:"required,password,nefield=OldPassword,alphanum"`
	OldPassword string `json:"old_password" validate:"required,password,nefield=NewPassword,alphanum"`
}
