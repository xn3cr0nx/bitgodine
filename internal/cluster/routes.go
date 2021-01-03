package cluster

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine/pkg/validator"
)

// Routes mounts all /clusters based routes on the main group
func Routes(g *echo.Group) {
	r := g.Group("/clusters", validator.JWT())

	r.GET("", getClusters)
	r.POST("", createCluster)
	r.GET("/:address", getClusterByAddress)
}

// getClusters godoc
// @ID get-clusters
//
// @Router /clusters [get]
// @Summary Get clusters list
// @Description get whole clusters list
// @Tags clusters
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
func getClusters(c echo.Context) error {
	type Query struct {
		Output bool `query:"output" validate:"omitempty"`
	}
	q := new(Query)
	if err := validator.Struct(&c, q); err != nil {
		return err
	}

	tags, err := GetClusters(&c, q.Output)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tags)
}

// createCluster godoc
// @ID create-cluster
//
// @Router /clusters [post]
// @Summary Create clusters
// @Description create a new cluster
// @Tags clusters
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param model body Model false "cluster model"
//
// @Success 200 {string} ok
// @Success 500 {string} string
func createCluster(c echo.Context) error {
	b := new(Model)
	if err := validator.Struct(&c, b); err != nil {
		return err
	}

	if err := CreateCluster(&c, b); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "Ok")
}

// getClustersByAddress godoc
// @ID get-clusters-by-address
//
// @Router /clusters/:address [get]
// @Summary Get clusters by address
// @Description get list of clusters based on address param
// @Tags clusters
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
func getClusterByAddress(c echo.Context) error {
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

	clusters, err := GetCluster(&c, address, q.Output)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, clusters)
}
