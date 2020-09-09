package trace

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/pkg/models"

	"github.com/xn3cr0nx/bitgodine/internal/abuse"
	"github.com/xn3cr0nx/bitgodine/internal/analysis"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine/internal/tag"
	"golang.org/x/sync/errgroup"
)

// Cluster struct to classify the address in a certain cluster
type Cluster struct {
	Type     string `json:"type,omitempty"`
	Message  string `json:"message,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Verified bool   `json:"verified,omitempty"`
}

// Trace between ouput and spending tx for tracing
type Trace struct {
	TxID string `json:"txid"`
	Next []Next `json:"next"`
}

// Next spending tx info
type Next struct {
	TxID     string    `json:"txid"`
	Receiver string    `json:"receiver"`
	Vout     uint32    `json:"vout"`
	Amount   float64   `json:"amount"`
	Weight   float64   `json:"weight"`
	Analysis string    `json:"analysis"`
	Clusters []Cluster `json:"clusters"`
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

			// find output with sought address
			vout := uint32(0)
			for o, out := range tx.Vout {
				if out.ScriptpubkeyAddress == address {
					vout = uint32(o)
				}
			}

			if err := followFlow(c, db, flow, tx, vout, 0, &lock); err != nil {
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
				clusters := []Cluster{}

				var g errgroup.Group
				g.Go(func() (err error) {
					tags, err := tag.GetTaggedClusterSet(c, tx.Vout[output].ScriptpubkeyAddress)
					if err != nil {
						if !strings.Contains(err.Error(), "cluster not found") {
							return err
						}
					}
					for _, tag := range tags {
						clusters = append(clusters, Cluster{
							Type:     tag.Type,
							Message:  tag.Message + " " + tag.Link,
							Nickname: tag.Nickname,
							Verified: tag.Verified,
						})
					}
					return
				})
				g.Go(func() (err error) {
					abuses, err := abuse.GetAbusedClusterSet(c, tx.Vout[output].ScriptpubkeyAddress)
					if err != nil {
						if !strings.Contains(err.Error(), "cluster not found") {
							return err
						}
					}
					for _, abuse := range abuses {
						clusters = append(clusters, Cluster{
							Type:     "abuse",
							Message:  abuse.Description,
							Nickname: abuse.Abuser,
							Verified: false,
						})
					}
					return
				})
				if err = g.Wait(); err != nil {
					return err
				}

				localNext = append(localNext, Next{
					TxID:     spending.TxID,
					Vout:     output,
					Receiver: tx.Vout[output].ScriptpubkeyAddress,
					Amount:   satToBtc(tx.Vout[output].Value),
					Weight:   percentage,
					Analysis: fmt.Sprintf("%b", mask[0]),
					Clusters: clusters,
				})
				err = followFlow(c, db, flow, spending, output, depth+1, lock)
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
