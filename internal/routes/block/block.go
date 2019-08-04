package block

import (
	"net/http"
	"strconv"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/models"
	"github.com/xn3cr0nx/bitgodine_code/internal/routes/tx"
	"github.com/xn3cr0nx/bitgodine_code/pkg/validator"

	"github.com/labstack/echo/v4"
)

func BlockToModel(b dgraph.Block) models.Block {
	return models.Block{
		ID:         b.Hash,
		Height:     uint32(b.Height),
		Version:    uint8(b.Version),
		Timestamp:  b.Time,
		Bits:       b.Bits,
		Nonce:      b.Nonce,
		MerkleRoot: b.MerkleRoot,
		TxCount:    len(b.Transactions),
		// Size:              ,
		// Weight:            ,
		Previousblockhash: b.PrevBlock,
	}
}

// Routes mounts all /block, /blocks and /block-height based routes on the main group
func Routes(g *echo.Group) *echo.Group {
	g.GET("/block-height/:height", func(c echo.Context) error {
		height, err := strconv.Atoi(c.Param("height"))
		if err != nil {
			return err
		}
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(height, "required,numeric,gte=0"); err != nil {
			return err
		}
		b, err := dgraph.GetBlockFromHeight(int32(height))
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		return c.JSON(http.StatusOK, b.Hash)
	})

	r := g.Group("/block")

	r.GET("/:hash", func(c echo.Context) error {
		hash := c.Param("hash")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required,testing"); err != nil {
			return err
		}
		b, err := dgraph.GetBlockFromHash(hash)
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		res := BlockToModel(b)
		return c.JSON(http.StatusOK, res)
	})

	// TODO: check if block in the best chain
	// r.GET("/:hash/status", func(c echo.Context) error {
	// 	hash := c.Param("hash")
	// 	if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required,testing"); err != nil {
	// 		return err
	// 	}
	// 	b, err := dgraph.GetBlockFromHash(hash)
	// 	if err != nil {
	// 		if err.Error() == "Block not found" {
	// 			return echo.NewHTTPError(http.StatusNotFound, err)
	// 		}
	// 		return err
	// 	}
	// 	res := BlockToModel(b)
	// 	return c.JSON(http.StatusOK, res)
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
		b, err := dgraph.GetBlockFromHash(hash)
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		var txs []models.Tx
		for i := start; i < 25+start; i++ {
			if i > len(b.Transactions)-1 {
				break
			}
			txs = append(txs, tx.TxToModel(b.Transactions[i], b.Height, b.Hash, b.Time))
		}
		return c.JSON(http.StatusOK, txs)
	})

	r.GET("/:hash/txids", func(c echo.Context) error {
		hash := c.Param("hash")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required,testing"); err != nil {
			return err
		}
		b, err := dgraph.GetBlockFromHash(hash)
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		var txids []string
		for _, tx := range b.Transactions {
			txids = append(txids, tx.Hash)
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
		blocks, err := dgraph.GetBlockFromHeightRange(int32(start), 10)
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		var res []models.Block
		for _, b := range blocks {
			res = append(res, BlockToModel(b))
		}
		return c.JSON(http.StatusOK, res)
	})

	s.GET("/tip/height", func(c echo.Context) error {
		b, err := dgraph.LastBlock()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, b.Height)
	})

	s.GET("/tip/hash", func(c echo.Context) error {
		b, err := dgraph.LastBlock()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, b.Hash)
	})

	return r
}
