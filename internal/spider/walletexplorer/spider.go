// Package walletexplorer Spider for address tags resource from walletexplorer.com
package walletexplorer

import (
	"fmt"

	"github.com/gocolly/colly/v2"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine/internal/tag"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
	"github.com/xn3cr0nx/bitgodine/pkg/postgres"
)

// Spider creates a web spider with predefined url to crawl
type Spider struct {
	crawler *colly.Collector
	target  string
	pg      *postgres.Pg
}

// NewSpider instance new spider object
func NewSpider(pg *postgres.Pg) Spider {
	return Spider{
		crawler: colly.NewCollector(colly.Async(true), colly.CacheDir(viper.GetString("spider.cache"))),
		target:  viper.GetString("spider.walletexplorer.url"),
		pg:      pg,
	}
}

// Sync visits the spider's target and extract address tags
func (s *Spider) Sync() (err error) {
	// last := &tag.Model{Type: category}
	// s.pg.DB.Where(last).First(last)
	// signedMessages, err := s.ExtractAddresses(strings.Join([]string{s.target, category}, "/"), category, last)
	_, err = s.ExtractAddresses(s.target)
	if err != nil {
		return
	}
	return nil
}

// ExtractAddresses returnes the list of signed messages
func (s *Spider) ExtractAddresses(target string) (tags []tag.Model, err error) {
	// NOT CURRENTLY IMPLEMENTED
	return
	var reached bool

	s.crawler.OnRequest(func(r *colly.Request) {
		logger.Info("Walletexplorer spider", "Visiting page", logger.Params{"url": r.URL.String()})
	})

	s.crawler.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		if reached {
			return
		}
		// tag := tag.Model{
		// 	Address:  e.ChildText("td:nth-child(1)"),
		// 	Message:  e.ChildText("td:nth-child(3)"),
		// 	Link:     e.ChildText("td:nth-child(4)"),
		// 	Type:     category,
		// 	Verified: verified,
		// }
		// if tag.Address == last.Address && tag.Link == last.Link && tag.Verified == last.Verified {
		// 	reached = true
		// 	return
		// }
		e.ForEachWithBreak("td ul", func(_ int, ce *colly.HTMLElement) bool {
			ce.ForEachWithBreak("li", func(_ int, le *colly.HTMLElement) bool {
				return true
			})
			return true
		})

		// tags = append(tags, tag)
	})

	// s.NavigateNextPage(target, reached)

	if err = s.crawler.Visit(target); err != nil {
		return
	}
	s.crawler.Wait()

	return
}

// NavigateNextPage scrapes pages incrementally extracting next page url from pages bar
func (s *Spider) NavigateNextPage(target string, reached bool) {
	s.crawler.OnHTML("div.container", func(e *colly.HTMLElement) {
		if reached {
			return
		}
		e.ForEachWithBreak("nav ul", func(_ int, ce *colly.HTMLElement) bool {
			var nth int
			ce.ForEachWithBreak("li", func(index int, le *colly.HTMLElement) bool {
				if le.Text == ">" {
					nth = index
					return false
				}
				return true
			})
			e.Request.Visit(target + e.ChildAttr(fmt.Sprintf("li:nth-child(%d) a", nth+1), "href"))
			return false
		})
	})
}
