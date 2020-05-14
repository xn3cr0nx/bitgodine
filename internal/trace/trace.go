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
	Weight   float64 `json:"weight"`
	Analysis string  `json:"analysis"`
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
		for _, output := range tx.Vout {
			if output.ScriptpubkeyAddress == address {
				spending, err := db.GetFollowingTx(tx.TxID, output.Index)
				if err != nil {
					return nil, err
				}
				if err := followFlow(c, db, flow, spending); err != nil {
					return nil, err
				}

			}
		}

	}

	fmt.Println("address occurences", occurences)
	fmt.Println("final flow", flow)

	return flow, nil
}

func followFlow(c *echo.Context, db storage.DB, flow *Flow, tx models.Tx) (err error) {
	changes, err := analysis.AnalyzeTx(c, tx.TxID, heuristics.FromListToMask(heuristics.List()), "reliability")
	if err != nil {
		if err.Error() == "Not feasible transaction" {
			return nil
		}
		return
	}
	likelihood, err := analysis.MajorityVotingOutput(changes.(heuristics.Map))
	if err != nil {
		if err.Error() == "Not feasible transaction" {
			return nil
		}
		return err
	}

	for output, percentages := range likelihood {
		spending, e := db.GetFollowingTx(tx.TxID, output)
		if e != nil {
			if e.Error() == "Key not found" {
				continue
			}
			return e
		}
		for mask, percentage := range percentages {
			flow.Traces[tx.TxID] = Trace{
				TxID:     tx.TxID,
				Vout:     output,
				Receiver: tx.Vout[output].ScriptpubkeyAddress,
				Amount:   satToBtc(tx.Vout[output].Value),
				Next:     spending.TxID,
				Weight:   percentage,
				Analysis: fmt.Sprintf("%b", mask[0]),
			}
			err := followFlow(c, db, flow, spending)
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
