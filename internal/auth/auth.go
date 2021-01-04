package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/jwt"
	"github.com/xn3cr0nx/bitgodine/internal/password"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/internal/user"
)

// Service interface exports available methods for user service
type Service interface {
	Login(body *LoginBody) (*LoginResp, error)
	Signup(body *SignupBody) (*SignupResp, error)
	GenerateAPIKey(ID, email string) (string, error)
	ChangePassword(ID, oldPassword, newPassword string) error
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

// Login verifies user exists and returns user data and authentication token
func (s *service) Login(body *LoginBody) (*LoginResp, error) {
	userService := user.NewService(s.Repository)
	user, err := userService.GetUserByEmail(body.Email)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, echo.ErrNotFound)
	}

	if !password.Verify(user.Password, body.Password) {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, echo.ErrUnauthorized)
	}

	t, err := jwt.NewToken(user.ID.String(), user.Email, time.Hour*24)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, echo.ErrValidatorNotRegistered)
	}

	lastLogin, err := userService.NewLogin(user.ID.String())
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, echo.ErrValidatorNotRegistered)
	}

	userResp := UserResp{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		LastLogin: lastLogin,
		IsActive:  user.IsActive,
		Lang:      user.Lang,
		IsBlocked: user.IsBlocked,
		APIKeys:   user.APIKeys,
	}
	resp := &LoginResp{
		Token: t,
		User:  userResp,
	}

	return resp, nil
}

// Signup create a new user
func (s *service) Signup(body *SignupBody) (*SignupResp, error) {
	verified := false
	if viper.GetBool("server.auth.directRegistration") {
		verified = true
	}

	email := strings.ToLower(body.Email)

	u := &user.Model{
		Email:     email,
		Password:  body.Password,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Username:  body.Username,
		Lang:      "en",
		IsActive:  verified,
	}

	// TODO: save uuid

	userService := user.NewService(s.Repository)
	if err := userService.CreateUser(u); err != nil {
		return nil, err
	}

	// token := random.String(32)
	// _, err = e.Redis.Set(token, u.ID.String(), 3600*time.Second).Result()
	// if err != nil {
	// 	log.LogError(c, action, err, nil)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }
	// url := fmt.Sprintf("%s/verification?key=%s", strings.Join([]string{viper.GetString("http.host"), viper.GetString("http.port")}, ":"), token)
	// emailNonce := random.String(32)
	// name := strings.Join([]string{u.FirstName, u.LastName}, " ")
	// resp, err := mailer.UserVerification(name, u.Email, url, c)
	// if err != nil || (resp != nil && resp.StatusCode >= 400) {
	// 	log.LogError(c, action, err, nil)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }
	// fields := map[string]interface{}{
	// 	"MAILER_FUNCTION": "UserVerification",
	// 	"name":            name,
	// 	"email":           u.Email,
	// 	"url":             url,
	// }
	// _, err = e.Redis.HMSet("email_nonce:"+emailNonce, fields).Result()
	// if err != nil {
	// 	log.LogError(c, action, err, nil)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }
	// _, err = e.Redis.Expire("email_nonce:"+emailNonce, 600*time.Second).Result()
	// if err != nil {
	// 	log.LogError(c, action, err, nil)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	// c.JSON(http.StatusOK, gin.H{
	// 	"success":     true,
	// 	"message":     "Signed up, check your email",
	// 	"email_nonce": emailNonce,
	// })
	return &SignupResp{"Check your email"}, nil
}

// GenerateAPIKey saves a new long term api key for the user
func (s *service) GenerateAPIKey(ID, email string) (token string, err error) {
	token, err = jwt.NewToken(ID, email, 99999999)
	if err != nil {
		return
	}

	userService := user.NewService(s.Repository)
	if err = userService.NewAPIKey(ID, token); err != nil {
		return
	}

	return
}

// ChangePassword changes user password
func (s *service) ChangePassword(ID, oldPassword, newPassword string) (err error) {
	userService := user.NewService(s.Repository)
	user, err := userService.GetUser(ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, echo.ErrNotFound)
	}

	if !password.Verify(user.Password, oldPassword) {
		return echo.NewHTTPError(http.StatusUnauthorized, echo.ErrUnauthorized)
	}

	user.Password, err = password.Hash(newPassword)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, echo.ErrValidatorNotRegistered)
	}
	err = userService.UpdateUser(ID, user)

	return
}
