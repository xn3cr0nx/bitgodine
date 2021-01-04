package abuse

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine/pkg/validator"
)

// Routes mounts all /abuse based routes on the main group
func Routes(g *echo.Group, s Service) {
	r := g.Group("/abuses", validator.JWT())

	r.GET("", getAbuses(s))
	r.POST("", createAbuse(s))
	r.GET("/:address", getAbusesByAddress(s))
	r.GET("/cluster/:address", getAbusedCluster(s))
	r.GET("/cluster/:address/set", getAbusedClusterSet(s))
}

// getAbuses godoc
// @ID get-abuses
//
// @Router /abuses [get]
// @Summary Get abuses list
// @Description get whole abuses list
// @Tags abuses
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param output query boolean false "Print output table"
//
// @Success 200 {array} Model
// @Success 500 {string} string
func getAbuses(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		type Query struct {
			Output bool `query:"output" validate:"omitempty"`
		}
		q := new(Query)
		if err := validator.Struct(&c, q); err != nil {
			return err
		}

		abuses, err := s.GetAbuses(q.Output)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, abuses)
	}
}

// createAbuse godoc
// @ID create-abuses
//
// @Router /abuses [post]
// @Summary Create abuses
// @Description create a new abuses
// @Tags abuses
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param model body Model false "abuse model"
//
// @Success 200 {string} ok
// @Success 500 {string} string
func createAbuse(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		b := new(Model)
		if err := validator.Struct(&c, b); err != nil {
			return err
		}

		if err := s.CreateAbuse(b); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, "Ok")
	}
}

// getAbusesByAddress godoc
// @ID get-abuses-by-address
//
// @Router /abuses/:address [get]
// @Summary Get abuse by address
// @Description get list of abuses based on address param
// @Tags abuses
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param address path string true "address"
// @Param output query boolean false "print table list"
//
// @Success 200 {array} Model
// @Success 500 {string} string
func getAbusesByAddress(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}

		type Query struct {
			Output bool `query:"output" validate:"omitempty"`
		}
		q := new(Query)
		if err := validator.Struct(&c, q); err != nil {
			return err
		}

		abuses, err := s.GetAbuse(address, q.Output)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, abuses)
	}
}

// getAbusedCluster godoc
// @ID get-abused-cluster
//
// @Router /abuses/cluster/:address [get]
// @Summary Get abused cluster
// @Description get list of abused clusters related address
// @Tags abuses
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param address path string true "address"
//
// @Success 200 {array} AbusedCluster
// @Success 500 {string} string
func getAbusedCluster(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}

		clusters, err := s.GetAbusedCluster(address)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, clusters)
	}
}

// getAbusedClusterSet godoc
// @ID get-abused-cluster-set
//
// @Router /abuses/cluster/:address/set [get]
// @Summary Get abused cluster set
// @Description get list of clusters related to address
// @Tags abuses
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param address path string true "address"
//
// @Success 200 {array} cluster.Model
// @Success 500 {string} string
func getAbusedClusterSet(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		address := c.Param("address")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
			return err
		}

		clusters, err := s.GetAbusedClusterSet(address)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, clusters)
	}
}
