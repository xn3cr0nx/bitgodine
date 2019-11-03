package validator

import (
	"net/http"

	"github.com/labstack/echo/v4"
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
func (cv *CustomValidator) Validate(i interface{}) (err error) {
	if err = cv.validator.Struct(i); err != nil {
		err = echo.NewHTTPError(http.StatusBadRequest, err)
	}
	return
}

// Var apply validation on passed variable
func (cv *CustomValidator) Var(i interface{}, tag string) (err error) {
	if err = cv.validator.Var(i, tag); err != nil {
		err = echo.NewHTTPError(http.StatusBadRequest, err)
	}
	return
}

// RegisterCustomValidations enhance the validator with a set of custom validators
func (cv *CustomValidator) RegisterCustomValidations() {
	_ = cv.validator.RegisterValidation("testing", TestingValidator)
	_ = cv.validator.RegisterValidation("jwt", JWTValidator)
}
