package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/xn3cr0nx/bitgodine/internal/jwt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

// RecaptchaResponse Google Recaptcha API response
type RecaptchaResponse struct {
	Success        bool      `json:"success"`
	ChallengeTs    time.Time `json:"Challenge_ts,omitempty"` // timestamp of the challenge load (ISO format yyyy-MM-dd'T'HH:mm:ssZZ)
	Hostname       string    `json:"hostname,omitempty"`     // the hostname of the site where the reCAPTCHA was solved
	ApkPackageName string    `json:"apk_package_name,omitempty"`
	ErrorCodes     []string  `json:"error_codes,omitempty"` // optional
}

// JWT middleware checks the user role is authorized to query the route
func JWT() func(echo.HandlerFunc) echo.HandlerFunc {
	if !viper.GetBool("auth.enabled") {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}
	return middleware.JWTWithConfig(jwt.Config())
}

// Recaptcha middleware to validate recaptcha input
func Recaptcha() func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !viper.GetBool("recaptcha.enabled") {
				return next(c)
			}

			type Body struct {
				Email         string `json:"email" validate:"required,email"`
				Password      string `json:"password" validate:"required,password,alphanum"`
				Lang          string `json:"lang,omitempty" validate:"omitempty,oneof=en ru de"`
				Role          string `json:"role,omitempty" validate:"omitempty,oneof=admin support-admin user accountant admin-kyc"`
				Promo         string `json:"promo,omitempty" validate:"omitempty,alphanum"`
				ReferrerToken string `json:"referrerToken,omitempty" validate:"omitempty,alphanum"`
				Captcha       string `json:"captcha,omitempty" validate:"omitempty"` // TODO: validate string with symbols too
				IsAndroid     bool   `json:"isAndroid"`
			}
			b := new(Body)
			if err := Struct(&c, b); err != nil {
				return err
			}
			c.Set("body", b)

			var secret string
			if b.IsAndroid {
				secret = viper.GetString("recaptcha.androidRecaptchaSecretKey")
			} else {
				secret = viper.GetString("recaptcha.recaptchaSecretKey")
			}

			remoteAddr, _, _ := net.SplitHostPort(c.Request().RemoteAddr)
			resp, err := http.Get(fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s&remoteip=%s", secret, b.Captcha, remoteAddr))
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			var recaptcha RecaptchaResponse
			if err := json.Unmarshal(body, &recaptcha); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err)
			}
			if recaptcha.Success == false {
				return echo.NewHTTPError(http.StatusBadRequest, errors.New(recaptcha.ErrorCodes[0]))
			}

			return next(c)
		}
	}
}
