package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/xn3cr0nx/bitgodine/internal/httpx"
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
	if !viper.GetBool("server.auth.enabled") {
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
			if !viper.GetBool("server.auth.recaptcha.enabled") {
				return next(c)
			}

			captcha := c.Request().Header[viper.GetString("auth.captchaKey")]
			secret := viper.GetString("recaptcha.recaptchaSecretKey")

			remoteAddr, _, _ := net.SplitHostPort(c.Request().RemoteAddr)
			resp, err := httpx.GET(fmt.Sprintf("https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s&remoteip=%s", secret, captcha, remoteAddr), nil)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err)
			}

			var recaptcha RecaptchaResponse
			if err := json.Unmarshal([]byte(resp), &recaptcha); err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err)
			}
			if recaptcha.Success == false {
				return echo.NewHTTPError(http.StatusBadRequest, errors.New(recaptcha.ErrorCodes[0]))
			}

			return next(c)
		}
	}
}
