// Package blockchain Spider for address tags resource from blockchain.com
package blockchain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/xn3cr0nx/bitgodine/internal/tag"
	"github.com/xn3cr0nx/bitgodine/pkg/postgres"
)

// Spider creates a web spider with predefined url to crawl
type Spider struct {
	crawler *colly.Collector
	target  string
	pg      *postgres.Pg
}

// Sync visits the spider's target and extract address tags
func (s *Spider) Sync() (err error) {
	var tags []tag.Model
	s.crawler.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	s.crawler.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		img := e.ChildAttr("img", "src")
		verified := false
		if img == "/Resources/green_tick.png" {
			verified = true
		}

		tags = append(tags, tag.Model{
			Address:  e.ChildText("td:nth-child(1)"),
			Message:  e.ChildText("td:nth-child(2)"),
			Link:     e.ChildText("td:nth-child(3)"),
			Verified: verified,
			Type:     "blockchain.com",
		})
	})

	s.crawler.OnHTML("li.next:not(.disabled) a", func(e *colly.HTMLElement) {
		next := e.Attr("href")
		if next != "" {
			split := strings.Split(next, "&")
			if len(split) == 1 {
				err = errors.New("Error in parsing next link")
			}
			e.Request.Visit(s.target + "&" + split[1])
		}
	})
	if err = s.crawler.Visit(s.target); err != nil {
		return
	}

	s.crawler.Wait()

	for _, t := range tags {
		if res := s.pg.DB.Create(&t); res.Error != nil {
			return res.Error
		}
	}

	return
}

// func (s *Spider) exportDB(data []tag.Tag) error {
// 	for _, t := range data {
// 		if t.Address == "" {
// 			continue
// 		}
// 		// if res := s.Pg.DB.Create(&tag.Tag{Address: tag.Address}); res.Error != nil {
// 		// 	return res.Error
// 		// }
// 		if res := s.pg.DB.Create(&t); res.Error != nil {
// 			return res.Error
// 		}
// 		logger.Info("Spider", fmt.Sprintf("Tag %s address %s correctly stored", t.Tag, t.Address), logger.Params{})
// 	}
// 	return nil
// }

// // exportCSV exports data to csv
// func exportCSV(data []tag.Tag, output string) error {
// 	if _, err := os.Stat(viper.GetString("tag.output")); os.IsNotExist(err) {
// 		os.Mkdir(viper.GetString("tag.output"), 0777)
// 	}

// 	filepath := fmt.Sprintf("%s/%s.csv", viper.GetString("tag.output"), output)
// 	var file *os.File
// 	if _, err := os.Stat(filepath); err == nil {
// 		file, err = os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
// 		if err != nil {
// 			return err
// 		}
// 	} else if os.IsNotExist(err) {
// 		file, err = os.Create(filepath)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	defer file.Close()
// 	writer := csv.NewWriter(file)
// 	defer writer.Flush()
// 	for _, tag := range data {
// 		if tag.Address == "" {
// 			continue
// 		}
// 		writer.Write([]string{tag.Address, tag.Tag, tag.Meta, strconv.FormatBool(tag.Verified)})
// 	}

// 	return nil
// }
