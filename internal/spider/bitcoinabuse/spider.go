// Package bitcoinabuse Spider for address tags resource from bitcoinabuse.com
package bitcoinabuse

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_server/internal/abuse"
	chttp "github.com/xn3cr0nx/bitgodine_server/internal/http"
	"github.com/xn3cr0nx/bitgodine_server/pkg/postgres"
)

// Spider creates a web spider with predefined url to crawl
type Spider struct {
	target string
	pg     *postgres.Pg
}

// NewSpider instance new spider object
func NewSpider(pg *postgres.Pg) Spider {
	target := fmt.Sprintf("%s/%s?api_token=%s", viper.GetString("spider.bitcoinabuse.url"), viper.GetString("spider.bitcoinabuse.period"), viper.GetString("spider.bitcoinabuse.api"))
	return Spider{
		target,
		pg,
	}
}

// updatePeriod replace period url param
func (s *Spider) updatePeriod(period string) {
	s.target = strings.Replace(s.target, viper.GetString("spider.bitcoinabuse.period"), period, 1)
}

// Sync visits the spider's target and extract address tags
func (s *Spider) Sync() (err error) {
	var count int
	s.pg.DB.Model(&abuse.Model{}).Count(&count)
	fmt.Println("count abuses", count)
	if count <= viper.GetInt("bitcoinabuse.reported_abuses") {
		s.updatePeriod("forever")
	}

	logger.Info("Spider", "Fetching bitcoinabuse reports", logger.Params{"target": s.target})
	resp, err := http.Get(s.target)
	if err != nil {
		return err
	}
	body, err := chttp.ParseResponse(resp)
	if err != nil {
		return err
	}

	logger.Info("Spider", "Parsing bitcoinabuse new resource", logger.Params{"length": len(body)})

	body = strings.ReplaceAll(body, ",\"\\\",", ",\\,")
	body = strings.ReplaceAll(body, "\\\",", "\",")
	body = strings.ReplaceAll(body, "\\\"", "")

	r := csv.NewReader(strings.NewReader(body))
	r.LazyQuotes = true
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	newResource := false
	for _, r := range records[1:] {
		id, e := strconv.Atoi(r[0])
		if e != nil {
			return e
		}
		created, e := time.Parse("2006-01-02 15:04:05", r[8])
		if e != nil {
			return e
		}
		m := gorm.Model{
			ID:        uint(id),
			CreatedAt: created,
		}
		a := &abuse.Model{
			Model:           m,
			Address:         r[1],
			AbuseTypeID:     r[2],
			AbuseTypeOther:  r[3],
			Abuser:          r[4],
			Description:     r[5],
			FromCountry:     r[6],
			FromCountryCode: r[7],
		}
		if count != 0 {
			if res := s.pg.DB.First(&abuse.Model{}, a.ID); res.Error == nil {
				continue
			}
		}
		newResource = true
		if res := s.pg.DB.Create(a); res.Error != nil {
			if strings.Contains(res.Error.Error(), "invalid byte sequence") {
				continue
			}
			return res.Error
		}
	}
	if !newResource {
		logger.Info("Spider", "No new abuse in resource", logger.Params{})
	}

	return
}
