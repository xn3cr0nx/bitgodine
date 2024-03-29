package tx

import (
	"errors"
	"net/http"

	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/pkg/validator"

	"github.com/labstack/echo/v4"
)

// Routes mounts all /tx based routes on the main group
func Routes(g *echo.Group, s Service) {
	r := g.Group("/tx", validator.JWT())
	r.GET("/:txid", txID(s))
	r.GET("/:txid/status", txIDStatus(s))

	// TODO: generate btcutil block and return hex conversion
	// r.GET("/:txid/hex", func(c echo.Context) error {
	// 	txid := c.Param("txid")
	// 	if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required"); err != nil {
	// 		return err
	// 	}
	// db := c.Get("db")
	// t, err := db.(kv.DB).GetTx(txid)
	// 	if err != nil {
	// 		if errors.Is(err, errorx.ErrKeyNotFound) {
	// 			err = echo.NewHTTPError(http.StatusNotFound, err)
	// 		}
	// 		return err
	// 	}
	// 	return c.JSON(http.StatusOK, fmt.Sprintf("%X", t))
	// })

	// TODO: returns something like this
	// {"block_height":142765,"merkle":["e4b7dc58ff92d7dc12429c13d2b3f55b498e25276c49ca607c2da4701570219e","92c70fb36d67e5cb391f7b4ebbbd1517c5530829c1e746e2f56dfff3f91b6cd5","baeb3d1f777f9314fdc9c4358abb5b2b96f47420255688a29e9ff2354a7c3f31","9e3465fa50ab32eff60d969827fb9a508d5bbe04fcd5de5eb8651cabeabf0e13","5352038e4e4f9325126faff1ecc8273dec78d1122bbf454b23078b73a8049e49","770c13987c512869ca926f4c84f3f1ca030750e69f5024afa478921eab88111b"],"pos":0}
	// r.GET("/:txid/merkle-proof", func(c echo.Context) error {
	//}

	// TODO: retrieve spent output
	// r.GET("/tx/:txid/outspend/:vout", func(c echo.Context) error {
	// 	txid := c.Param("txid")
	// 	if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required"); err != nil {
	// 		return err
	// 	}
	// 	vout := c.Param("vout")
	// 	if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required,numeric,gte=0"); err != nil {
	// 		return err
	// 	}
	// db := c.Get("db")
	// t, err := db.(kv.DB).GetTx(txid)
	// 	if err != nil {
	// 		if errors.Is(err, errorx.ErrKeyNotFound) {
	// 			err = echo.NewHTTPError(http.StatusNotFound, err)
	// 		}
	// 		return err
	// 	}
	// 	block, err := dgraph.GetTxBlock(txid)
	// 	if err != nil {
	// 		if errors.Is(err, errorx.ErrKeyNotFound) {
	// 			return echo.NewHTTPError(http.StatusNotFound, err)
	// 		}
	// 		return err
	// 	}
	// 	transaction := TxToModel(t, block.Height, block.Hash, block.Time)

	// 	// return c.JSON(http.StatusOK, fmt.Sprintf("%X", t))
	// })

	r.GET("/:txid/outspend/:vout", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})

	r.GET("/:txid/outspends", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})

	// // TODO: receive hex and broadcast tx
	// r.POST("", func(c echo.Context) error {
	// 	return c.JSON(http.StatusOK, "OK")
	// })
}

// txID godoc
// @ID tx-id
//
// @Router /tx/{txid} [get]
// @Summary Tx from id
// @Description get transaction from id
// @Tags tx
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param txid path string true "Transaction id"
//
// @Success 200 {object} Tx
// @Success 500 {string} string
func txID(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		txid := c.Param("txid")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required"); err != nil {
			return err
		}
		t, err := s.GetFromHash(txid)
		if err != nil {
			if errors.Is(err, errorx.ErrKeyNotFound) {
				err = echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		return c.JSON(http.StatusOK, t)
	}
}

// txIDStatus godoc
// @ID tx-id-status
//
// @Router /tx/{txid}/status [get]
// @Summary Tx status from id
// @Description get transaction's status from id
// @Tags tx
//
// @Security ApiKeyAuth
//
// @Accept  json
// @Produce  json
//
// @Param txid path string true "Transaction id"
//
// @Success 200 {object} Tx
// @Success 500 {string} string
func txIDStatus(s Service) func(echo.Context) error {
	return func(c echo.Context) error {
		txid := c.Param("txid")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required"); err != nil {
			return err
		}
		t, err := s.GetFromHash(txid)
		if err != nil {
			if errors.Is(err, errorx.ErrKeyNotFound) {
				err = echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		return c.JSON(http.StatusOK, t.Status)
	}
}
