// This file is part of Aguaxaca.
// Copyright (C) 2025 Arnaud Berthomier.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package collector

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/gocolly/colly/v2"
)

// BaseDomain is the Nitter domain where we searth tweets.
// const defaultBaseDomain = "https://nitter.net"
const defaultBaseDomain = "https://nitter.net"

// OutputDir is the directory where images will be saved.
const defaultOutputDir = "./images"

type NitterCollector struct {
	BaseDomain string
	OutputDir  string
	Account    string
}

// NewNitterCollector builds a default Nitter collector / scraper.
func NewNitterCollector(account string) *NitterCollector {
	// TODO: make this configurable
	return &NitterCollector{
		BaseDomain: defaultBaseDomain,
		OutputDir:  defaultOutputDir,
		Account:    account,
	}
}

// DownloadImages scrapes images from a Nitter HTML timeline, and
// returns a list of paths when images where downloaded.
func (nc *NitterCollector) DownloadImages() ([]string, error) {
	c := nc.getFirefoxCollector()
	files := []string{}

	err := os.Mkdir(nc.OutputDir, 0750)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	// Lookup timeline items
	c.OnHTML(".timeline-item", func(e *colly.HTMLElement) {
		txt := e.ChildText("*")
		// Ignore tweets that don't contain the correct hashtag.
		if !strings.Contains(txt, "HoyLlegaElAgua") {
			return
		}

		e.ForEach(".attachments a.still-image", func(i int, a *colly.HTMLElement) {
			imgURL := nc.BaseDomain + a.Attr("href")
			log.Printf("Found image at: \"%s\"\n", imgURL)

			file, err := nc.downloadImage(imgURL)
			if err != nil {
				log.Printf("Error downloading \"%s\": %s\n", imgURL, err)
				return
			}
			files = append(files, file)
		})
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("HTTP error %d for %s: %v\n", r.StatusCode, r.Request.URL, err)
	})

	c.Visit(nc.BaseDomain + "/" + nc.Account)
	log.Println("HTML scraper: waiting for pending requests")
	c.Wait()
	log.Println("HTML scraper: done.")
	return files, nil
}

// Try to download image once.
func (nc *NitterCollector) downloadImage(url string) (string, error) {
	fileName := path.Base(url)
	dest := path.Join(nc.OutputDir, fileName)

	// This check avoids an unnecessary copy from Colly's local cache.
	if fileExists(dest) {
		log.Printf("Skipped image: %s is in cache\n", url)
		return dest, nil
	}
	log.Printf("Downloading %s to: %s\n", url, dest)

	c := nc.getFirefoxCollector()
	c.OnResponse(func(r *colly.Response) {
		if strings.Contains(r.Headers.Get("Content-Type"), "image") {
			fileName := path.Base(r.Request.URL.Path)
			if err := r.Save(dest); err != nil {
				log.Printf("Error saving file %s: %s\n", fileName, err)
			}
		}
	})
	c.Visit(url)
	c.Wait()

	return dest, nil
}

// Get a rude colly collector that mimicks Firefox headers.
func (nc *NitterCollector) getFirefoxCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:140.0) Gecko/20100101 Firefox/140.0"),

		// Don't use that with a public Nitter instance.
		colly.IgnoreRobotsTxt(),

		// TODO: CacheExpiration isn't automatic in colly v2.2.0.
		colly.CacheDir("cache"),
	)
	c.OnRequest(func(r *colly.Request) {
		log.Println("GET", r.URL)
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br, zstd")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Pragma", "no-cache")
		r.Headers.Set("Priority", "u=0, i")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("DNT", "1")
		r.Headers.Set("Referer", nc.BaseDomain)
	})
	return c
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
