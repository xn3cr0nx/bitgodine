package trace

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/xn3cr0nx/bitgodine/internal/address"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"
	"github.com/xn3cr0nx/bitgodine/internal/storage/db/postgres"
	"github.com/xn3cr0nx/bitgodine/internal/storage/kv"
	"github.com/xn3cr0nx/bitgodine/internal/tx"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"

	"github.com/xn3cr0nx/bitgodine/internal/abuse"
	"github.com/xn3cr0nx/bitgodine/internal/analysis"
	"github.com/xn3cr0nx/bitgodine/internal/heuristics"
	"github.com/xn3cr0nx/bitgodine/internal/tag"
	"golang.org/x/sync/errgroup"
)

// Service interface exports available methods for block service
type Service interface {
	TraceAddress(address string, limit, skip int) (tracing *Flow, err error)
	followFlow(flow map[string]Trace, transaction tx.Tx, vout uint32, depth int, lock *sync.RWMutex) (err error)
}

type service struct {
	Repository *postgres.Pg
	Kv         kv.DB
	Cache      *cache.Cache
}

// NewService instantiates a new Service layer for customer
func NewService(r *postgres.Pg, k kv.DB, c *cache.Cache) *service {
	return &service{
		Repository: r,
		Kv:         k,
		Cache:      c,
	}
}

func (s *service) TraceAddress(addr string, limit, skip int) (tracing *Flow, err error) {
	fmt.Println("Tracing address", addr)
	occurences, err := address.NewService(s.Kv, s.Cache).GetOccurences(addr)
	if err != nil {
		return
	}
	tracing = &Flow{
		Traces:     make([]map[string]Trace, limit),
		Occurences: occurences,
	}

	var g errgroup.Group
	txService := tx.NewService(s.Kv, s.Cache)
	for i, occurence := range occurences {
		if i < limit*skip || i > skip*limit+(limit-1) {
			continue
		}

		occ := occurence
		flow := make(map[string]Trace)

		lock := sync.RWMutex{}
		index := i

		g.Go(func() error {
			occ = strings.Replace(occ, addr+"_", "", 1)
			transaction, err := txService.GetFromHash(occ)
			if err != nil {
				return err
			}

			// find output with sought address
			vout := uint32(0)
			for o, out := range transaction.Vout {
				if out.ScriptpubkeyAddress == addr {
					vout = uint32(o)
				}
			}

			if err := s.followFlow(flow, transaction, vout, 0, &lock); err != nil {
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

func (s *service) followFlow(flow map[string]Trace, transaction tx.Tx, vout uint32, depth int, lock *sync.RWMutex) (err error) {
	analysisService := analysis.NewService(s.Kv, s.Cache)
	changes, err := analysisService.AnalyzeTx(transaction.TxID, heuristics.FromListToMask(heuristics.List()), "reliability")
	if err != nil {
		if errors.Is(err, analysis.ErrUnfeasibleTx) {
			lock.Lock()
			flow[fmt.Sprintf("%s:%d", transaction.TxID, vout)] = Trace{
				TxID: transaction.TxID,
				Next: []Next{},
			}
			lock.Unlock()
			return nil
		}
		return
	}
	likelihood, err := analysis.MajorityVotingOutput(changes.(heuristics.Map))
	if err != nil {
		if errors.Is(err, analysis.ErrUnfeasibleTx) {
			lock.Lock()
			flow[fmt.Sprintf("%s:%d", transaction.TxID, vout)] = Trace{
				TxID: transaction.TxID,
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
	txService := tx.NewService(s.Kv, s.Cache)
	for out, perc := range likelihood {
		output := out
		percentages := perc
		g.Go(func() error {
			spending, e := txService.GetSpendingFromHash(transaction.TxID, output)
			if e != nil {
				if errors.Is(err, errorx.ErrKeyNotFound) {
					return nil
				}
				return e
			}
			var localNext []Next
			tagService := tag.NewService(s.Repository, s.Cache)
			for mask, percentage := range percentages {
				clusters := []Cluster{}

				var g errgroup.Group
				g.Go(func() (err error) {
					tags, err := tagService.GetTaggedClusterSet(transaction.Vout[output].ScriptpubkeyAddress)
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
					abuseService := abuse.NewService(s.Repository, s.Cache)
					abuses, err := abuseService.GetAbusedClusterSet(transaction.Vout[output].ScriptpubkeyAddress)
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
					Receiver: transaction.Vout[output].ScriptpubkeyAddress,
					Amount:   satToBtc(transaction.Vout[output].Value),
					Weight:   percentage,
					Analysis: fmt.Sprintf("%b", mask[0]),
					Clusters: clusters,
				})
				err = s.followFlow(flow, spending, output, depth+1, lock)
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
	flow[fmt.Sprintf("%s:%d", transaction.TxID, vout)] = Trace{
		TxID: transaction.TxID,
		Next: next,
	}
	lock.Unlock()

	return nil
}

func satToBtc(amount int64) float64 {
	return float64(amount) * math.Pow(10, -8)
}
