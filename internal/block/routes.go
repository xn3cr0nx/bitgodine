package block

import (
	"net/http"
	"strconv"

	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_server/pkg/validator"

	"github.com/labstack/echo/v4"
)

// Routes mounts all /block, /blocks and /block-height based routes on the main group
func Routes(g *echo.Group) *echo.Group {
	g.GET("/block-height/:height", func(c echo.Context) error {
		height, err := strconv.Atoi(c.Param("height"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(height, "numeric,gte=0"); err != nil {
			return err
		}

		db := c.Get("db")
		b, err := db.(storage.DB).GetBlockFromHeight(int32(height))
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return err
		}
		return c.JSON(http.StatusOK, b)
	})

	r := g.Group("/block")

	r.GET("/:hash", func(c echo.Context) error {
		hash := c.Param("hash")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required,testing"); err != nil {
			return err
		}
		db := c.Get("db")
		b, err := db.(storage.DB).GetBlockFromHash(hash)
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		return c.JSON(http.StatusOK, b)
	})

	// TODO: check if block in the best chain
	// r.GET("/:hash/status", func(c echo.Context) error {
	// 	hash := c.Param("hash")
	// 	if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required,testing"); err != nil {
	// 		return err
	// 	}
	// 	db := c.Get("db")
	// 	b, err := db.(storage.DB).GetBlockFromHash(hash)
	// 	if err != nil {
	// 		if err.Error() == "Block not found" {
	// 			return echo.NewHTTPError(http.StatusNotFound, err)
	// 		}
	// 		return err
	// 	}
	// 	return c.JSON(http.StatusOK, b)
	// })

	r.GET("/:hash/txs/:start_index", func(c echo.Context) error {
		hash := c.Param("hash")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required,testing"); err != nil {
			return err
		}
		start, err := strconv.Atoi(c.Param("start_index"))
		if err != nil {
			return err
		}
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(start, "omitempty,numeric,gte=0"); err != nil {
			return err
		}
		db := c.Get("db")
		b, err := db.(storage.DB).GetBlockFromHash(hash)
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		// TODO: fetch txs
		// var txs []models.Tx
		var txs []string
		for i := start; i < 25+start; i++ {
			if i > len(b.Transactions)-1 {
				break
			}
			// txs = append(txs, tx.TxToModel(b.Transactions[i], b.Height, b.ID, b.Timestamp))
			txs = append(txs, b.Transactions[i])
		}
		return c.JSON(http.StatusOK, txs)
	})

	r.GET("/:hash/txids", func(c echo.Context) error {
		hash := c.Param("hash")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required,testing"); err != nil {
			return err
		}
		db := c.Get("db")
		b, err := db.(storage.DB).GetBlockFromHash(hash)
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		var txids []string
		for _, tx := range b.Transactions {
			txids = append(txids, tx)
		}
		return c.JSON(http.StatusOK, txids)
	})

	s := g.Group("/blocks")
	s.GET("/:start_height", func(c echo.Context) error {
		start, err := strconv.Atoi(c.Param("start_index"))
		if err != nil {
			return err
		}
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(start, "omitempty,numeric,gte=0"); err != nil {
			return err
		}
		db := c.Get("db")
		blocks, err := db.(storage.DB).GetBlockFromHeightRange(int32(start), 10)
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		var res []models.Block
		for _, b := range blocks {
			res = append(res, b)
		}
		return c.JSON(http.StatusOK, res)
	})

	s.GET("/tip/height", func(c echo.Context) error {
		db := c.Get("db")
		b, err := db.(storage.DB).LastBlock()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, b.Height)
	})

	s.GET("/tip/hash", func(c echo.Context) error {
		db := c.Get("db")
		b, err := db.(storage.DB).LastBlock()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, b.ID)
	})

	return r
}
