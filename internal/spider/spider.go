package spider

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Spider creates a web spider with predefined url to crawl
type Spider struct {
	*colly.Collector
}

type Tag struct {
	address  string
	tag      string
	meta     string
	verified bool
}

// Sync visits the passed URL and extract address tags
func (s *Spider) Sync(URL, output string, finish chan bool) {
	var tags []Tag
	s.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	s.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		img := e.ChildAttr("img", "src")
		verified := false
		if img == "/Resources/green_tick.png" {
			verified = true
		}

		tags = append(tags, Tag{
			address:  e.ChildText("td:nth-child(1)"),
			tag:      e.ChildText("td:nth-child(2)"),
			meta:     e.ChildText("td:nth-child(3)"),
			verified: verified,
		})
	})

	s.OnHTML("li.next:not(.disabled) a", func(e *colly.HTMLElement) {
		next := e.Attr("href")
		if next != "" {
			split := strings.Split(next, "&")
			if len(split) == 1 {
				logger.Error("Spider", errors.New("Error in parsing next link"), logger.Params{})
				os.Exit(1)
			}
			e.Request.Visit(URL + "&" + split[1])
		}
	})
	if err := s.Visit(URL); err != nil {
		logger.Error("Spider", err, logger.Params{})
		os.Exit(1)
	}

	s.Wait()
	if err := exportCSV(tags, output); err != nil {
		logger.Error("Spider", err, logger.Params{})
		os.Exit(1)
	}
	finish <- true
}

// exportCSV exports data to csv
func exportCSV(data []Tag, output string) error {
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
		writer.Write([]string{tag.address, tag.tag, tag.meta, strconv.FormatBool(tag.verified)})
	}

	return nil
}
