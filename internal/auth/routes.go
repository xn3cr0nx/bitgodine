package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine/internal/jwt"
	"github.com/xn3cr0nx/bitgodine/pkg/validator"
)

// Routes mounts auth routes on the main group
func Routes(g *echo.Group, s Service) {
	g.POST("/login", login(s))
	g.POST("/signup", signup(s), validator.Recaptcha())
	// g.GET("/generate-api-key", generateAPIKey(s), validator.JWT())
	g.GET("/generate-api-key", generateAPIKey(s), validator.JWT())
	g.POST("/change-password", changePassword(s), validator.JWT())

	// r.GET("/activate/:token", func(c echo.Context) error {
	// 	token := c.Param("token")
	// 	tags := []string{"required"}
	// 	if viper.GetBool("auth.enabled") {
	// 		tags = append(tags, "jwt")
	// 	}
	// 	if err := c.Echo().Validator.(*validator.CustomValidator).Var(token, strings.Join(tags, ",")); err != nil {
	// 		return err
	// 	}
	// 	resp, err := Activate(c)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return c.JSON(http.StatusOK, resp)
	// })

	// r.POST("/forgot-confirm", func(c echo.Context) error {
	// 	type Body struct {
	// 		Token    string `json:"token" validate:"required"`
	// 		Password string `json:"password" validate:"required,password,alphanum"`
	// 	}
	// 	b := new(Body)
	// 	if err := validator.Struct(&c, b); err != nil {
	// 		return err
	// 	}
	// 	resp, err := ForgotConfirm(c, b)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return c.JSON(http.StatusOK, resp)
	// })

	// r.POST("/resend", func(c echo.Context) error {
	// 	type Body struct {
	// 		Email string `json:"email,omitempty" validate:"required,email"`
	// 	}
	// 	b := new(Body)
	// 	if err := validator.Struct(&c, b); err != nil {
	// 		return err
	// 	}

	// 	res, err := chttp.POST(c.Request().RequestURI, b, routes.ProxyAuth(&c))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	resp := new(models.Response)
	// 	if err := json.Unmarshal([]byte(res), resp); err != nil {
	// 		return err
	// 	}
	// 	return c.JSON(http.StatusOK, resp)
	// })

	// r.GET("/email-api-key", func(c echo.Context) error {
	// 	res, err := chttp.GET(c.Request().RequestURI, routes.ProxyAuth(&c))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	resp := new(models.Response)
	// 	if err := json.Unmarshal([]byte(res), resp); err != nil {
	// 		return err
	// 	}
	// 	return c.JSON(http.StatusOK, resp)
	// })

	// r.POST("/revoke-api-key", func(c echo.Context) error {
	// 	type Body struct {
	// 		APIKeyIndex uint32 `json:"apiKeyIndex" validate:"required,numeric,gte=0"`
	// 	}
	// 	b := new(Body)
	// 	if err := validator.Struct(&c, b); err != nil {
	// 		return err
	// 	}

	// 	res, err := chttp.POST(c.Request().RequestURI, b, routes.ProxyAuth(&c))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	resp := new(models.Response)
	// 	if err := json.Unmarshal([]byte(res), resp); err != nil {
	// 		return err
	// 	}
	// 	return c.JSON(http.StatusOK, resp)
	// })

	// r.POST("/forgot", func(c echo.Context) error {
	// 	b := new(Login)
	// 	if err := validator.Struct(&c, b); err != nil {
	// 		return err
	// 	}

	// 	res, err := chttp.POST(c.Request().RequestURI, b, routes.ProxyAuth(&c))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	resp := new(models.Response)
	// 	if err := json.Unmarshal([]byte(res), resp); err != nil {
	// 		return err
	// 	}
	// 	return c.JSON(http.StatusOK, resp)
	// })
}

// login godoc
// @ID login
//
// @Router /login [post]
// @Summary Login
// @Description Authenticate
// @Tags auth
//
// @Accept  json
// @Produce  json
//
// @Param login body LoginBody true "login body"
//
// @Success 200 {object} LoginResp
// @Failure 400 {string} string
// @Failure 500 {string} string
func login(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		b := new(LoginBody)
		if err := validator.Struct(&c, b); err != nil {
			return err
		}

		resp, err := s.Login(b)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, resp)
	}
}

// signup godoc
// @ID signup
//
// @Router /signup [post]
// @Summary Signup
// @Description Authenticate
// @Tags auth
//
// @Accept  json
// @Produce  json
//
// @Param signup body SignupBody true "signup body"
//
// @Success 200 {object} SignupResp
// @Failure 400 {string} string
// @Failure 500 {string} string
func signup(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		b := new(SignupBody)
		if err := validator.Struct(&c, b); err != nil {
			return err
		}

		resp, err := s.Signup(b)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, resp)
	}
}

// generateAPIKey godoc
// @ID generateAPIKey
//
// @Router /generate-api-key [get]
// @Summary Generate Api Key
// @Description Generate and new api key for the user
// @Tags auth
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
func generateAPIKey(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		claims, err := jwt.Decode(c.Get("user"))
		if err != nil {
			return nil
		}

		resp, err := s.GenerateAPIKey(claims.ID, claims.Email)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, resp)
	}
}

// changePassword godoc
// @ID changePassword
//
// @Router /chage-password [post]
// @Summary Change Password
// @Description Authenticate
// @Tags auth
//
// @Accept  json
// @Produce  json
//
// @Param changePassword body ChangePasswordBody true "change password body"
//
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
func changePassword(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		b := new(ChangePasswordBody)
		if err := validator.Struct(&c, b); err != nil {
			return err
		}

		claims, err := jwt.Decode(c.Get("user"))
		if err != nil {
			return nil
		}

		if err := s.ChangePassword(claims.ID, b.OldPassword, b.NewPassword); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, "ok")
	}
}
