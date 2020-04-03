package trace

import (
	"fmt"
	"math"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_server/internal/analysis"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
)

// Trace between ouput and spending tx for tracing
type Trace struct {
	TxID     string  `json:"txid"`
	Receiver string  `json:"receiver"`
	Vout     uint32  `json:"vout"`
	Amount   float64 `json:"amount"`
	Next     string  `json:"next"`
}

// Flow list of maps creating monetary flow
type Flow struct {
	Traces map[string]Trace `json:"traces"`
}

func traceAddress(c *echo.Context, address string) (*Flow, error) {
	fmt.Println("Tracing address", address)
	flow := &Flow{
		Traces: make(map[string]Trace),
	}

	db := (*c).Get("db").(storage.DB)
	occurences, err := db.GetAddressOccurences(address)
	if err != nil {
		return nil, err
	}

	for _, occurence := range occurences {
		occurence = strings.Replace(occurence, address+"_", "", 1)
		tx, err := db.GetTx(occurence)
		if err != nil {
			return nil, err
		}

		err = followFlow(c, db, flow, tx, address)
		if err != nil {
			if err.Error() == "No effective analysis" {
				continue
			}
			return nil, err
		}
	}

	fmt.Println("address occurences", occurences)
	fmt.Println("final flow", flow)

	return flow, nil
}

func followFlow(c *echo.Context, db storage.DB, flow *Flow, tx models.Tx, address string) error {
	if address == "" {
		changes, e := analysis.AnalyzeTx(c, tx.TxID, heuristics.FromListToMask(heuristics.List()), "reliability")
		if e != nil {
			if e.Error() == "Not feasible transaction" {
				return nil
			}
			return e
		}
		output, e := analysis.ExtractLikelihoodOutput(changes.(heuristics.Map))
		if e != nil {
			if e.Error() == "No effective analysis" {
				flow.Traces[tx.TxID] = Trace{
					TxID:     tx.TxID,
					Next:     "?",
					Receiver: "?",
				}
				return nil
			}
			return e
		}
		address = tx.Vout[int(output)].ScriptpubkeyAddress
		fmt.Println("NEXT ADDRESS", address)
	}

	for _, output := range tx.Vout {
		if output.ScriptpubkeyAddress == address {
			spending, e := db.GetFollowingTx(tx.TxID, output.Index)
			if e != nil {
				if e.Error() == "Key not found" {
					continue
				}
				return e
			}

			flow.Traces[tx.TxID] = Trace{
				TxID:     tx.TxID,
				Vout:     output.Index,
				Receiver: output.ScriptpubkeyAddress,
				Amount:   satToBtc(output.Value),
				Next:     spending.TxID,
			}
			err := followFlow(c, db, flow, spending, "")
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func satToBtc(amount int64) float64 {
	return float64(amount) * math.Pow(10, -8)
}
