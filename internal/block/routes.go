package block

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/pkg/validator"

	"github.com/labstack/echo/v4"
)

// Routes mounts all /block, /blocks and /block-height based routes on the main group
func Routes(g *echo.Group, s Service) {
	g.GET("/block-height/:height", blockHeight(s), validator.JWT())

	r := g.Group("/block", validator.JWT())
	r.GET("/:hash", blockHash(s))
	// r.GET("/:hash/status", blockStatus)
	r.GET("/:hash/txs/:start_index", blockHashTxs(s))
	r.GET("/:hash/txids", blockHashTxIDs(s))

	b := g.Group("/blocks", validator.JWT())
	b.GET("/tip/height", tipHeight(s))
	b.GET("/tip/hash", tipHash(s))
	b.GET("/:start_height", blocksHeight(s))
}

// blockHeight godoc
// @ID block-height
//
// @Router /block-height/{height} [get]
// @Summary Block from height
// @Description get block from height
// @Tags block
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param id path int true "Block height"
//
// @Success 200 {object} BlockOut
// @Success 500 {string} string
func blockHeight(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		height, err := strconv.Atoi(c.Param("height"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest)
		}
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(height, "numeric,gte=0"); err != nil {
			return err
		}

		b, err := s.GetFromHeight(int32(height))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, b)
	}
}

// TODO: check if block in the best chain
// func blockStatus(s Service) func(echo.Context) error {
// 	return func(c echo.Context) error {
// hash := c.Param("hash")
// 	if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required,testing"); err != nil {
// 		return err
// 	}
// 	db := c.Get("db")
// 	b, err := db.(kv.DB).GetFromHash(hash)
// 	if err != nil {
// 		if errors.Is(err, errorx.ErrKeyNotFound) {
// 			return echo.NewHTTPError(http.StatusNotFound, err)
// 		}
// 		return err
// 	}
// 	return c.JSON(http.StatusOK, b)
// 	}
// }

// blockHash godoc
// @ID block-hash
//
// @Router /block/{hash} [get]
// @Summary Block from hash
// @Description get block from hash
// @Tags block
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param hash path string true "Block hash"
// @Success 200 {object} BlockOut
// @Success 500 {string} string
func blockHash(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		hash := c.Param("hash")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required,testing"); err != nil {
			return err
		}
		b, err := s.GetFromHashWithTxs(hash)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, b)
	}
}

// blockHashTxs godoc
// @ID block-hash-txs
//
// @Router /block/{hash}/txs/{start_index} [get]
// @Summary Block transactions
// @Description get block transactions from hash starting from index
// @Tags block
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param hash path string true "Block hash"
// @Param start_index path int false "Transactions starting index"
// @Success 200 {array} string
// @Success 500 {string} string
func blockHashTxs(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		hash := c.Param("hash")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required"); err != nil {
			return err
		}
		start, err := strconv.Atoi(c.Param("start_index"))
		if err != nil {
			return err
		}
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(start, "omitempty,numeric,gte=0"); err != nil {
			return err
		}
		b, err := s.GetFromHash(hash)
		if err != nil {
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
	}
}

// blockHashTxIDs godoc
// @ID block-hash-tx-ids
//
// @Router /block/{hash}/txids [get]
// @Summary Block transaction ids
// @Description get block transaction ids from hash
// @Tags block
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param hash path string true "Block hash"
// @Success 200 {array} string
// @Success 500 {string} string
func blockHashTxIDs(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		hash := c.Param("hash")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(hash, "required"); err != nil {
			return err
		}
		b, err := s.GetFromHash(hash)
		if err != nil {
			return err
		}

		var txids []string
		for _, tx := range b.Transactions {
			txids = append(txids, tx)
		}
		return c.JSON(http.StatusOK, txids)
	}
}

// blocksHeight godoc
// @ID blocks-height
//
// @Router /blocks/{start_height} [get]
// @Summary Blocks from height
// @Description get blocks starting from height (10 by default)
// @Tags blocks
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param start_height path int true "Starting height"
// @Success 200 {array} BlockOut
// @Success 500 {string} string
func blocksHeight(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		start, err := strconv.Atoi(c.Param("start_height"))
		if err != nil {
			return err
		}
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(start, "omitempty,numeric,gte=0"); err != nil {
			return err
		}
		blocks, err := s.GetFromHeightRange(int32(start), 10)
		if err != nil {
			if errors.Is(err, errorx.ErrKeyNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		var res []BlockOut
		for _, b := range blocks {
			res = append(res, b)
		}
		return c.JSON(http.StatusOK, res)
	}
}

// tipHeight godoc
// @ID tip-height
//
// @Router /blocks/tip/height [get]
// @Summary Tip height
// @Description get tip block height
// @Tags blocks
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Success 200 {number} int
// @Success 500 {string} string
func tipHeight(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		b, err := s.GetLast()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, b.Height)
	}
}

// tipHash godoc
// @ID tip-hash
//
// @Router /blocks/tip/hash [get]
// @Summary Tip hash
// @Description get tip block hash
// @Tags blocks
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Success 200 {string} hash
// @Success 500 {string} string
func tipHash(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		b, err := s.GetLast()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, b.ID)
	}
}
