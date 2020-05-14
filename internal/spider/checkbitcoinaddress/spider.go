// Package checkbitcoinaddress Spider for address tags resource from checkbitcoinaddress.com
package checkbitcoinaddress

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_parser/pkg/logger"
	"github.com/xn3cr0nx/bitgodine_server/internal/tag"
	"github.com/xn3cr0nx/bitgodine_server/pkg/postgres"
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
		crawler: colly.NewCollector(colly.Async(true), colly.CacheDir(viper.GetString("cache"))),
		target:  viper.GetString("checkbitcoinaddress.url"),
		pg:      pg,
	}
}

// Sync visits the spider's target and extract address tags
func (s *Spider) Sync() (err error) {
	category := "signed-messages"
	last := &tag.Model{Type: category}
	s.pg.DB.Where(last).First(last)
	signedMessages, err := s.ExtractSignedMessages(strings.Join([]string{s.target, category}, "/"), category, last)
	if err != nil {
		return
	}

	category = "submitted-links"
	last = &tag.Model{Type: category}
	s.pg.DB.Where(last).First(last)
	s.crawler = colly.NewCollector(colly.Async(true), colly.CacheDir(viper.GetString("cache")))
	submittedLinks, err := s.ExtractSubmittedLinks(strings.Join([]string{s.target, category}, "/"), category, last)
	if err != nil {
		return
	}

	category = "bitcoin-otc-profiles"
	last = &tag.Model{Type: category}
	s.pg.DB.Where(last).First(last)
	s.crawler = colly.NewCollector(colly.Async(true), colly.CacheDir(viper.GetString("cache")))
	otcProfiles, err := s.ExtractOTCProfiles(strings.Join([]string{s.target, category}, "/"), category, last)
	if err != nil {
		return
	}

	category = "forum-profiles"
	last = &tag.Model{Type: category}
	s.pg.DB.Where(last).First(last)
	s.crawler = colly.NewCollector(colly.Async(true), colly.CacheDir(viper.GetString("cache")))
	forumProfiles, err := s.ExtractForumProfiles(strings.Join([]string{s.target, category}, "/"), category, last)
	if err != nil {
		return
	}

	var tags []tag.Model
	tags = append(tags, signedMessages...)
	tags = append(tags, submittedLinks...)
	tags = append(tags, otcProfiles...)
	tags = append(tags, forumProfiles...)

	for _, t := range tags {
		if res := s.pg.DB.Create(&t); res.Error != nil {
			return res.Error
		}
	}

	return
}

// ExtractSignedMessages returnes the list of signed messages
func (s *Spider) ExtractSignedMessages(target, category string, last *tag.Model) (tags []tag.Model, err error) {
	var reached bool

	s.crawler.OnRequest(func(r *colly.Request) {
		logger.Info("Checkbitcoinaddress spider", "Visiting page", logger.Params{"url": r.URL.String()})
	})

	s.crawler.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		if reached {
			return
		}
		img := e.ChildAttr("img", "src")
		verified := false
		if img == "/i/tick.svg" {
			verified = true
		}
		tag := tag.Model{
			Address:  e.ChildText("td:nth-child(1)"),
			Message:  e.ChildText("td:nth-child(3)"),
			Link:     e.ChildText("td:nth-child(4)"),
			Type:     category,
			Verified: verified,
		}
		if tag.Address == last.Address && tag.Link == last.Link && tag.Verified == last.Verified {
			reached = true
			return
		}
		tags = append(tags, tag)
	})

	s.NavigateNextPage(target, reached)

	if err = s.crawler.Visit(target); err != nil {
		return
	}
	s.crawler.Wait()

	return
}

// ExtractSubmittedLinks returnes the list of submitted links
func (s *Spider) ExtractSubmittedLinks(target, category string, last *tag.Model) (tags []tag.Model, err error) {
	tags, err = s.ExtractSignedMessages(target, category, last)
	for _, t := range tags {
		t.Type = category
	}
	return
}

// ExtractOTCProfiles returnes the list of otc profiles
func (s *Spider) ExtractOTCProfiles(target, category string, last *tag.Model) (tags []tag.Model, err error) {
	tags, err = s.ExtractSignedMessages(target, category, last)
	for _, t := range tags {
		t.Type = category
		t.Nickname = t.Message
		t.Message = ""
	}
	return
}

// ExtractForumProfiles returnes the list of forum profiles
func (s *Spider) ExtractForumProfiles(target, category string, last *tag.Model) (tags []tag.Model, err error) {
	var reached bool

	s.crawler.OnRequest(func(r *colly.Request) {
		logger.Info("Checkbitcoinaddress spider", "Visiting page", logger.Params{"url": r.URL.String()})
	})

	s.crawler.OnHTML("div.row div div.profile-card div.card-body", func(e *colly.HTMLElement) {
		if reached {
			return
		}

		img := e.ChildText("div div strong")
		verified := false
		if img == " Verified Member" {
			verified = true
		}
		tag := tag.Model{
			Address:  e.ChildText("p.small a"),
			Nickname: e.ChildText("h2"),
			Link:     e.ChildAttr("p.quote-text a", "href"),
			Verified: verified,
			Type:     category,
		}
		if tag.Address == last.Address && tag.Link == last.Link && tag.Verified == last.Verified {
			reached = true
			return
		}
		tags = append(tags, tag)
	})

	s.NavigateNextPage(target, reached)

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
