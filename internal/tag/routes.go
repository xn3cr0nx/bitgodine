package tag

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine/pkg/validator"
)

// Routes mounts all /tags based routes on the main group
func Routes(g *echo.Group) {
	r := g.Group("/tags", validator.JWT())

	r.GET("", getTags)
	r.POST("", createTag)
	r.GET("/:address", getTagByAddress)
	r.GET("/cluster/:address", getTaggedClusterByAddress)
	r.GET("/cluster/:address/set", getTaggedClusterSetByAddress)
}

// getTags godoc
// @ID get-tags
//
// @Router /tags [get]
// @Summary Get tags list
// @Description get whole tags list
// @Tags tags
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
func getTags(c echo.Context) error {
	type Query struct {
		Output bool `query:"output" validate:"omitempty"`
	}
	q := new(Query)
	if err := validator.Struct(&c, q); err != nil {
		return err
	}

	tags, err := GetTags(&c, q.Output)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tags)
}

// createTag godoc
// @ID create-tags
//
// @Router /tags [post]
// @Summary Create tags
// @Description create a new tags
// @Tags tags
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param model body Model false "tag model"
//
// @Success 200 {string} ok
// @Success 500 {string} string
func createTag(c echo.Context) error {
	b := new(Model)
	if err := validator.Struct(&c, b); err != nil {
		return err
	}

	if err := CreateTag(&c, b); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "")
}

// getTagsByAddress godoc
// @ID get-tags-by-address
//
// @Router /tags/:address [get]
// @Summary Get tag by address
// @Description get list of tags based on address param
// @Tags tags
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
func getTagByAddress(c echo.Context) error {
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

	tags, err := GetTag(&c, address, q.Output)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tags)
}

// getTaggedClusterByAddress godoc
// @ID get-tagged-cluster-by-address
//
// @Router /tags/cluster/:address [get]
// @Summary Get tag by address
// @Description get list of tagged clusters based on address param
// @Tags tags
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param address path string true "address"
//
// @Success 200 {array} TaggedCluster
// @Success 500 {string} string
func getTaggedClusterByAddress(c echo.Context) error {
	address := c.Param("address")
	if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
		return err
	}

	clusters, err := GetTaggedCluster(&c, address)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, clusters)
}

// getTaggedClusterSetByAddress godoc
// @ID get-tagged-cluster-set-by-address
//
// @Router /tags/cluster/:address/set [get]
// @Summary Get tag by address
// @Description get list of tagged cluster set based on address param
// @Tags tags
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param address path string true "address"
//
// @Success 200 {array} Model
// @Success 500 {string} string
func getTaggedClusterSetByAddress(c echo.Context) error {
	address := c.Param("address")
	if err := c.Echo().Validator.(*validator.CustomValidator).Var(address, "required,btc_addr|btc_addr_bech32"); err != nil {
		return err
	}

	clusters, err := GetTaggedClusterSet(&c, address)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, clusters)
}
