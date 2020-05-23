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
	Traces map[string]Trace `json:"traces"`
}

func traceAddress(c *echo.Context, address string) (*Flow, error) {
	fmt.Println("Tracing address", address)
	flow := &Flow{
		Traces: make(map[string]Trace),
	}
	lock := sync.RWMutex{}

	db := (*c).Get("db").(storage.DB)
	occurences, err := db.GetAddressOccurences(address)
	if err != nil {
		return nil, err
	}
	fmt.Println("Address Occurences", occurences)

	var g errgroup.Group
	for _, occurence := range occurences {
		g.Go(func() error {
			occurence = strings.Replace(occurence, address+"_", "", 1)
			tx, err := db.GetTx(occurence)
			if err != nil {
				return err
			}
			if err := followFlow(c, db, flow, tx, &lock); err != nil {
				return err
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return nil, err
	}

	return flow, nil
}

func followFlow(c *echo.Context, db storage.DB, flow *Flow, tx models.Tx, lock *sync.RWMutex) (err error) {
	changes, err := analysis.AnalyzeTx(c, tx.TxID, heuristics.FromListToMask(heuristics.List()), "reliability")
	if err != nil {
		if err.Error() == "Not feasible transaction" {
			lock.Lock()
			flow.Traces[tx.TxID] = Trace{
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
			flow.Traces[tx.TxID] = Trace{
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
	for output, percentages := range likelihood {
		g.Go(func() error {
			spending, e := db.GetFollowingTx(tx.TxID, output)
			if e != nil {
				if e.Error() == "Key not found" {
					// continue
					return nil
				}
				return e
			}
			// var likely Next
			var localNext []Next
			for mask, percentage := range percentages {
				// if likely.TxID == "" || percentage > likely.Weight {
				// likely = Next{
				localNext = append(localNext, Next{
					TxID:     spending.TxID,
					Vout:     output,
					Receiver: tx.Vout[output].ScriptpubkeyAddress,
					Amount:   satToBtc(tx.Vout[output].Value),
					Weight:   percentage,
					Analysis: fmt.Sprintf("%b", mask[0]),
				})
				// }
				err := followFlow(c, db, flow, spending, lock)
				if err != nil {
					return err
				}
				// }
			}
			// next = append(next, likely)
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
	flow.Traces[tx.TxID] = Trace{
		TxID: tx.TxID,
		Next: next,
	}
	lock.Unlock()

	return nil
}

func satToBtc(amount int64) float64 {
	return float64(amount) * math.Pow(10, -8)
}
