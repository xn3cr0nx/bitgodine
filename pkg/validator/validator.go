package validator

import (
	"gopkg.in/go-playground/validator.v9"
)

// CustomValidator wrapper for validator
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator returnes new validator instance
func NewValidator() *CustomValidator {
	// return &CustomValidator{validator: validator.New()}
	v := &CustomValidator{validator: validator.New()}
	v.RegisterCustomValidations()
	return v
}

// Validate apply validation on passed interface
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// Var apply validation on passed variable
func (cv *CustomValidator) Var(i interface{}, tag string) error {
	return cv.validator.Var(i, tag)
}

// RegisterCustomValidations enhance the validator with a set of custom validators
func (cv *CustomValidator) RegisterCustomValidations() {
	_ = cv.validator.RegisterValidation("testing", TestingValidator)
	_ = cv.validator.RegisterValidation("jwt", JWTValidator)
}
