package trace

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/models"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/storage"
	"github.com/xn3cr0nx/bitgodine_server/internal/analysis"
	"github.com/xn3cr0nx/bitgodine_server/internal/heuristics"
	"golang.org/x/sync/errgroup"
)

// // Trace between ouput and spending tx for tracing
// type Trace struct {
// 	TxID     string  `json:"txid"`
// 	Receiver string  `json:"receiver"`
// 	Vout     uint32  `json:"vout"`
// 	Amount   float64 `json:"amount"`
// 	Next     string  `json:"next"`
// 	Weight   float64 `json:"weight"`
// 	Analysis string  `json:"analysis"`
// }

// Trace between ouput and spending tx for tracing
type Trace struct {
	TxID string `json:"txid"`
	Next []Next `json:"next"`
}

// Next spending tx info
type Next struct {
	TxID     string  `json:"txid"`
	Receiver string  `json:"receiver"`
	Vout     uint32  `json:"vout"`
	Amount   float64 `json:"amount"`
	Weight   float64 `json:"weight"`
	Analysis string  `json:"analysis"`
}

// Flow list of maps creating monetary flow
type Flow struct {
	Traces     []map[string]Trace `json:"traces"`
	Occurences []string           `json:"occurences"`
}

func traceAddress(c *echo.Context, address string, limit int, skip int) (tracing *Flow, err error) {
	fmt.Println("Tracing address", address)

	db := (*c).Get("db").(storage.DB)
	occurences, err := db.GetAddressOccurences(address)
	if err != nil {
		return
	}
	tracing = &Flow{
		Traces:     make([]map[string]Trace, limit),
		Occurences: occurences,
	}

	var g errgroup.Group
	for i, occurence := range occurences {
		if i < limit*skip || i > skip*limit+(limit-1) {
			continue
		}

		occ := occurence
		flow := make(map[string]Trace)

		lock := sync.RWMutex{}
		index := i

		g.Go(func() error {
			occ = strings.Replace(occ, address+"_", "", 1)
			tx, err := db.GetTx(occ)
			if err != nil {
				return err
			}
			if err := followFlow(c, db, flow, tx, 0, 0, &lock); err != nil {
				return err
			}
			tracing.Traces[index%limit] = flow
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return nil, err
	}

	return
}

func followFlow(c *echo.Context, db storage.DB, flow map[string]Trace, tx models.Tx, vout uint32, depth int, lock *sync.RWMutex) (err error) {
	changes, err := analysis.AnalyzeTx(c, tx.TxID, heuristics.FromListToMask(heuristics.List()), "reliability")
	if err != nil {
		if err.Error() == "Not feasible transaction" {
			lock.Lock()
			flow[fmt.Sprintf("%s:%d", tx.TxID, vout)] = Trace{
				TxID: tx.TxID,
				Next: []Next{},
			}
			lock.Unlock()
			return nil
		}
		return
	}
	likelihood, err := analysis.MajorityVotingOutput(changes.(heuristics.Map))
	if err != nil {
		if err.Error() == "Not feasible transaction" {
			lock.Lock()
			flow[fmt.Sprintf("%s:%d", tx.TxID, vout)] = Trace{
				TxID: tx.TxID,
				Next: []Next{},
			}
			lock.Unlock()
			return nil
		}
		return err
	}

	var g errgroup.Group
	var next []Next
	nextLock := sync.RWMutex{}
	for out, perc := range likelihood {
		output := out
		percentages := perc
		g.Go(func() error {
			spending, e := db.GetFollowingTx(tx.TxID, output)
			if e != nil {
				if e.Error() == "Key not found" {
					return nil
				}
				return e
			}
			var localNext []Next
			for mask, percentage := range percentages {
				localNext = append(localNext, Next{
					TxID:     spending.TxID,
					Vout:     output,
					Receiver: tx.Vout[output].ScriptpubkeyAddress,
					Amount:   satToBtc(tx.Vout[output].Value),
					Weight:   percentage,
					Analysis: fmt.Sprintf("%b", mask[0]),
				})
				err := followFlow(c, db, flow, spending, output, depth+1, lock)
				if err != nil {
					return err
				}
			}
			nextLock.Lock()
			next = append(next, localNext...)
			nextLock.Unlock()
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return
	}
	lock.Lock()
	flow[fmt.Sprintf("%s:%d", tx.TxID, vout)] = Trace{
		TxID: tx.TxID,
		Next: next,
	}
	lock.Unlock()

	return nil
}

func satToBtc(amount int64) float64 {
	return float64(amount) * math.Pow(10, -8)
}
