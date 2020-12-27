package jwt

import (
	"fmt"
	"time"

	token "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
)

// CustomClaims custom token object
type CustomClaims struct {
	ID string `json:"id"`
	token.StandardClaims
}

// Config returns custom JWTConfig object
func Config() middleware.JWTConfig {
	return middleware.JWTConfig{
		Claims:     &CustomClaims{},
		SigningKey: []byte(viper.GetString("auth.secret")),
		ContextKey: "token",
	}
}

// NewToken returns a new jwt token based on CustomClaims structure
func NewToken(id string, d time.Duration) (string, error) {
	claims := &CustomClaims{
		id,
		token.StandardClaims{
			ExpiresAt: time.Now().Add(d).Unix(),
		},
	}
	token := token.NewWithClaims(token.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(viper.GetString("auth.secret")))
	if err != nil {
		return "", err
	}
	return t, nil
}

// Validate returns an error is the token is invalid
func Validate(t string) error {
	// tk, err := token.Parse(t, func(tkn *token.Token) (interface{}, error) {
	// 	return []byte("AllYourBase"), nil
	// })
	tk, err := token.ParseWithClaims(t, &CustomClaims{}, func(*token.Token) (interface{}, error) {
		return []byte("Boh don't understand"), nil
	})
	if err != nil {
		fmt.Println("JWT VALIDATE ERROR", err)
		return err
	}
	fmt.Println("what out", tk)

	if tk.Valid {
		fmt.Println("You look nice today")
		return nil
	} else if ve, ok := err.(*token.ValidationError); ok {
		if ve.Errors&token.ValidationErrorMalformed != 0 {
			return fmt.Errorf("%w: That's not even a token", errorx.ErrInvalidArgument)
		} else if ve.Errors&(token.ValidationErrorExpired|token.ValidationErrorNotValidYet) != 0 { // Token is either expired or not active yet
			return fmt.Errorf("%w: Timing is everything", errorx.ErrInvalidArgument)
		} else {
			return fmt.Errorf("%w: %s", errorx.ErrUnknown, err.Error())
		}
	} else {
		return fmt.Errorf("%w: %s", errorx.ErrUnknown, err.Error())
	}
}

// Valid custom validation method for CustomClaims token object
func (c *CustomClaims) Valid() error {
	return nil
}
