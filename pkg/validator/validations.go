package validator

import (
	"regexp"

	"github.com/xn3cr0nx/bitgodine/internal/jwt"

	"github.com/go-playground/validator/v10"
)

// TestingValidator to test custom validators
func TestingValidator(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) > 5
}

// JWTValidator validates that the field contains a valid jwt token
func JWTValidator(fl validator.FieldLevel) bool {
	token := fl.Field().String()
	if err := jwt.Validate(token); err != nil {
		return false
	}
	return true
}

// LimitValidator to test custom validators
func LimitValidator(fl validator.FieldLevel) bool {
	return fl.Field().Int() < 500 && fl.Field().Int()%5 == 0
}

// PasswordValidator to test custom validators
func PasswordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	includesNumber, _ := regexp.MatchString(".*[0-9].*", password)
	includesUpper, _ := regexp.MatchString(".*[A-Z].*", password)
	return len(password) > 6 && includesNumber && includesUpper
}
