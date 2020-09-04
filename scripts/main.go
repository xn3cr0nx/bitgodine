package main

import (
	"archive/tar"
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

// Tag of tag struct with validation
type Tag struct {
	gorm.Model
	Address  string `json:"address" validate:"required,btc_addr|btc_addr_bech32" gorm:"size:64;index;not null"`
	Message  string `json:"message" validate:"required" gorm:"index;not null"`
	Nickname string `json:"nickname,omitempty" validate:"" gorm:"index;not null"`
	Type     string `json:"type,omitempty" validate:"" gorm:"index;not null"`
	Link     string `json:"link,omitempty" validate:""`
	Verified bool   `json:"verified,omitempty" validate:"" gorm:"default:false"`
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func untar(dst string, r io.Reader) error {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
			f.Close()
		}
	}
}

func downloadDataset() (err error) {
	out, err := os.Create("./Addresses.zip")
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get("https://polybox.ethz.ch/index.php/s/GUEFVnOEPrMxY2l/download")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.Print("Downloading addresses dataset...")
	err := downloadDataset()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Unzipping dataset...")
	err = unzip("./Addresses.zip", ".")
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Opening tar file...")
	f, err := os.Open("./Addresses.tar")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.Print("Untaring tar dataset ...")
	err = untar("./Addresses", f)
	if err != nil {
		log.Fatal(err)
	}

	sources := []string{"Exchanges_full_detailed.csv", "Gambling_full_detailed.csv", "Historic_full_detailed.csv", "Mining_full_detailed.csv", "Services_full_detailed.csv"}
	tags := make([]Tag, 0)

	for c, source := range sources {
		log.Print("Extracting data from ", source)
		f, err := os.Open("./Addresses/Addresses/" + source)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		lines, err := csv.NewReader(f).ReadAll()
		if err != nil {
			log.Fatal(err)
		}

		category := ""
		if c == 0 {
			category = "exchange"
		} else if c == 1 {
			category = "gambling"
		} else if c == 2 {
			category = "historic"
		} else if c == 3 {
			category = "mining"
		} else if c == 4 {
			category = "services"
		} else {
			os.Exit(-1)
		}

		for i, line := range lines {
			// header row
			if i == 0 {
				continue
			}

			if c == 0 {
				tags = append(tags, Tag{
					Type:     category,
					Address:  line[5],
					Nickname: line[4],
					Message:  line[2] + " " + line[3],
					Verified: true,
				})
			} else {
				tags = append(tags, Tag{
					Type:     category,
					Address:  line[4],
					Nickname: line[3],
					Message:  line[2],
					Verified: true,
				})
			}
		}
	}

	file, err := os.Create("walletexplorer.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	log.Print("Writing whole tags csv file")
	for _, tag := range tags {
		if tag.Verified {
			writer.Write([]string{tag.Address, tag.Message, tag.Nickname, tag.Type, tag.Link, "true"})
		} else {
			writer.Write([]string{tag.Address, tag.Message, tag.Nickname, tag.Type, tag.Link, "false"})
		}
	}
}
