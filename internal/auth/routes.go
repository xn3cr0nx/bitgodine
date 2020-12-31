package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine/pkg/validator"
)

// Routes mounts auth routes on the main group
func Routes(g *echo.Group, s Service) {
	g.POST("/login", login(s))

	g.POST("/signup", signup(s), validator.Recaptcha())

	// g.GET("/restricted", restricted, middleware.JWTWithConfig(jwt.Config()))
	g.GET("/restricted", restricted, validator.JWT())

	// r.POST("/change-password", func(c echo.Context) error {
	// 	type Body struct {
	// 		NewPassword string `json:"newPassword" validate:"required,password,nefield=OldPassword,alphanum"`
	// 		OldPassword string `json:"oldPassword" validate:"required,password,nefield=NewPassword,alphanum"`
	// 	}
	// 	b := new(Body)
	// 	if err := validator.Struct(&c, b); err != nil {
	// 		return err
	// 	}
	// 	resp, err := ChangePassword(c, b)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return c.JSON(http.StatusOK, resp)
	// })

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

	// r.POST("/generate-api-key", func(c echo.Context) error {
	// 	res, err := chttp.POST(c.Request().RequestURI, nil, routes.ProxyAuth(&c))
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
//
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
//
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

// restricted godoc
// @ID restricted
//
// @Router /restricted [get]
// @Summary restricted
// @Description Authenticate
// @Tags auth
//
// @Accept  json
// @Produce  json
//
//
// @Success 500 {string} string
func restricted(c echo.Context) error {
	return c.JSON(http.StatusOK, "ok")
}