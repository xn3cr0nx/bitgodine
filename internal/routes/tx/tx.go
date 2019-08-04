package tx

import (
	"net/http"
	"strings"
	"time"

	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/models"
	"github.com/xn3cr0nx/bitgodine_code/pkg/validator"

	"github.com/labstack/echo/v4"
)

func TxToModel(tx dgraph.Transaction, height int32, blockHash string, time time.Time) models.Tx {
	var inputs []models.Input
	for _, input := range tx.Inputs {
		inputs = append(inputs, models.Input{
			TxID:       input.Hash,
			Vout:       input.Vout,
			IsCoinbase: input.Hash != strings.Repeat("0", 64),
			Scriptsig:  input.SignatureScript,
			// ScriptsigAsm: ,
			// InnerRedeemscriptAsm: ,
			// InnerWitnessscriptAsm: ,
			// Sequence: ,
			// Witness: ,
			// Prevout: ,
			IsPegin: false,
			// Issuance: ,
		})
	}
	var outputs []models.Output
	for _, output := range tx.Outputs {
		outputs = append(outputs, models.Output{
			Scriptpubkey: output.PkScript,
			// ScriptpubkeyAsm: ,
			// ScriptpubkeyType: ,
			ScriptpubkeyAddress: output.Address,
			Value:               uint64(output.Value),
			// Valuecommitment: ,
			// Asset: ,
			// Pegout: ,
		})
	}
	return models.Tx{
		TxID:     tx.Hash,
		Version:  uint8(tx.Version),
		Locktime: tx.Locktime,
		Size:     -1,
		Weight:   -1,
		Fee:      -1,
		Vin:      inputs,
		Vout:     outputs,
		Status: models.Status{
			Confirmed:   true,
			BlockHeight: uint32(height),
			BlockHash:   blockHash,
			BlockTime:   time,
		},
	}
}

// Routes mounts all /tx based routes on the main group
func Routes(g *echo.Group) *echo.Group {
	r := g.Group("/tx")

	r.GET("/:txid", func(c echo.Context) error {
		txid := c.Param("txid")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required,testing"); err != nil {
			return err
		}
		t, err := dgraph.GetTx(txid)
		if err != nil {
			if err.Error() == "transaction not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		block, err := dgraph.GetTxBlock(txid)
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		res := TxToModel(t, block.Height, block.Hash, block.Time)
		return c.JSON(http.StatusOK, res)
	})

	r.GET("/:txid/status", func(c echo.Context) error {
		txid := c.Param("txid")
		if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required,testing"); err != nil {
			return err
		}
		t, err := dgraph.GetTx(txid)
		if err != nil {
			if err.Error() == "transaction not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		block, err := dgraph.GetTxBlock(txid)
		if err != nil {
			if err.Error() == "Block not found" {
				return echo.NewHTTPError(http.StatusNotFound, err)
			}
			return err
		}
		res := TxToModel(t, block.Height, block.Hash, block.Time)
		return c.JSON(http.StatusOK, res.Status)
	})

	// TODO: generate btcutil block nad return hex conversion
	// r.GET("/:txid/hex", func(c echo.Context) error {
	// 	txid := c.Param("txid")
	// 	if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required,testing"); err != nil {
	// 		return err
	// 	}
	// 	t, err := dgraph.GetTx(txid)
	// 	if err != nil {
	// 		if err.Error() == "transaction not found" {
	// 			return echo.NewHTTPError(http.StatusNotFound, err)
	// 		}
	// 		return err
	// 	}
	// 	return c.JSON(http.StatusOK, fmt.Sprintf("%X", t))
	// })

	// TODO: returnes something like this
	// {"block_height":142765,"merkle":["e4b7dc58ff92d7dc12429c13d2b3f55b498e25276c49ca607c2da4701570219e","92c70fb36d67e5cb391f7b4ebbbd1517c5530829c1e746e2f56dfff3f91b6cd5","baeb3d1f777f9314fdc9c4358abb5b2b96f47420255688a29e9ff2354a7c3f31","9e3465fa50ab32eff60d969827fb9a508d5bbe04fcd5de5eb8651cabeabf0e13","5352038e4e4f9325126faff1ecc8273dec78d1122bbf454b23078b73a8049e49","770c13987c512869ca926f4c84f3f1ca030750e69f5024afa478921eab88111b"],"pos":0}
	// r.GET("/:txid/merkle-proof", func(c echo.Context) error {
	//}

	// TODO: retrieve spent output
	// r.GET("/tx/:txid/outspend/:vout", func(c echo.Context) error {
	// 	txid := c.Param("txid")
	// 	if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required,testing"); err != nil {
	// 		return err
	// 	}
	// 	vout := c.Param("vout")
	// 	if err := c.Echo().Validator.(*validator.CustomValidator).Var(txid, "required,numeric,gte=0"); err != nil {
	// 		return err
	// 	}
	// 	t, err := dgraph.GetTx(txid)
	// 	if err != nil {
	// 		if err.Error() == "transaction not found" {
	// 			return echo.NewHTTPError(http.StatusNotFound, err)
	// 		}
	// 		return err
	// 	}
	// 	block, err := dgraph.GetTxBlock(txid)
	// 	if err != nil {
	// 		if err.Error() == "Block not found" {
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

	// TODO: receive hex and broadcast tx
	r.POST("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})

	return r
}