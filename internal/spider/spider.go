package spider

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/postgres"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Spider creates a web spider with predefined url to crawl
type Spider struct {
	Spider *colly.Collector
	Target string
	Pg     *postgres.Pg
}

// // Tag contains data about a tag instance
// type Tag struct {
// 	address  string
// 	tag      string
// 	meta     string
// 	verified bool
// }

// Sync visits the spider's target and extract address tags
func (s *Spider) Sync(output string, wg *sync.WaitGroup) {
	var tags []postgres.Tag
	s.Spider.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	s.Spider.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		img := e.ChildAttr("img", "src")
		verified := false
		if img == "/Resources/green_tick.png" {
			verified = true
		}

		tags = append(tags, postgres.Tag{
			Address:  e.ChildText("td:nth-child(1)"),
			Tag:      e.ChildText("td:nth-child(2)"),
			Meta:     e.ChildText("td:nth-child(3)"),
			Verified: verified,
		})
	})

	s.Spider.OnHTML("li.next:not(.disabled) a", func(e *colly.HTMLElement) {
		next := e.Attr("href")
		if next != "" {
			split := strings.Split(next, "&")
			if len(split) == 1 {
				logger.Error("Spider", errors.New("Error in parsing next link"), logger.Params{})
				os.Exit(1)
			}
			e.Request.Visit(s.Target + "&" + split[1])
		}
	})
	if err := s.Spider.Visit(s.Target); err != nil {
		logger.Error("Spider", err, logger.Params{})
		os.Exit(1)
	}

	s.Spider.Wait()

	switch viper.GetString("tag.dest") {
	case "db":
		if err := s.exportDB(tags); err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(1)
		}
	case "csv":
		if err := exportCSV(tags, output); err != nil {
			logger.Error("Spider", err, logger.Params{})
			os.Exit(1)
		}
	}

	wg.Done()
}

func (s *Spider) exportDB(data []postgres.Tag) error {
	for _, tag := range data {
		if tag.Address == "" {
			continue
		}
		if res := s.Pg.DB.Create(&postgres.Address{Address: tag.Address}); res.Error != nil {
			return res.Error
		}
		if res := s.Pg.DB.Create(&tag); res.Error != nil {
			return res.Error
		}
		logger.Info("Spider", fmt.Sprintf("Tag %s address %s correctly stored", tag.Tag, tag.Address), logger.Params{})
	}
	return nil
}

// exportCSV exports data to csv
func exportCSV(data []postgres.Tag, output string) error {
	if _, err := os.Stat(viper.GetString("tag.output")); os.IsNotExist(err) {
		os.Mkdir(viper.GetString("tag.output"), 0777)
	}

	filepath := fmt.Sprintf("%s/%s.csv", viper.GetString("tag.output"), output)
	var file *os.File
	if _, err := os.Stat(filepath); err == nil {
		file, err = os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		file, err = os.Create(filepath)
		if err != nil {
			return err
		}
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	for _, tag := range data {
		if tag.Address == "" {
			continue
		}
		writer.Write([]string{tag.Address, tag.Tag, tag.Meta, strconv.FormatBool(tag.Verified)})
	}

	return nil
}
