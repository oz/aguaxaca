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
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/gocolly/colly/v2"
)

// DefaultBaseDomain is the Nitter instance where we scrape tweets.
const DefaultBaseDomain = "https://nitter.net"

// defaultDownloadDir is where images will be saved.
const DefaultDownloadDir = "./images"

type fileExistsError struct {
	name string
}

func (e *fileExistsError) Error() string {
	return fmt.Sprintf("File already exist: %s", e.name)
}

type NitterCollector struct {
	BaseDomain  string
	DownloadDir string
	Account     string
	Log         *slog.Logger
}

// NewNitterCollector builds a default Nitter collector / scraper.
// This is alright with a local, or private, Nitter instance. Better
// alternatives would be using Nitter's RSS feeds, or get permission
// to use a public instance. Do NOT use a public instance without
// permission from the instance's operators.
func NewNitterCollector(account string) *NitterCollector {
	return &NitterCollector{
		Account:     account,
		BaseDomain:  DefaultBaseDomain,
		DownloadDir: DefaultDownloadDir,
		Log:         slog.Default(),
	}
}

// DownloadImages scrapes images from a Nitter HTML timeline, and
// returns a list of paths when images where downloaded.
func (nc *NitterCollector) DownloadImages() ([]string, error) {
	c := nc.getFirefoxCollector()
	files := []string{}

	if err := os.Mkdir(nc.DownloadDir, 0750); err != nil && !os.IsExist(err) {
		return nil, err
	}

	// Collect all scraping errors
	var err error = nil

	c.OnHTML("title", func(e *colly.HTMLElement) {
		if strings.Index(e.Text, "Maintenance") == 0 {
			err = errors.Join(fmt.Errorf("server unavailable (maintenance)"))
		}
	})

	// Lookup timeline items
	c.OnHTML(".timeline-item", func(e *colly.HTMLElement) {
		txt := e.ChildText("*")
		// Ignore tweets that don't contain the correct hashtag.
		if !strings.Contains(txt, "HoyLlegaElAgua") {
			return
		}

		e.ForEach(".attachments a.still-image", func(i int, a *colly.HTMLElement) {
			imgURL := nc.BaseDomain + a.Attr("href")
			nc.Log.Info("found image", "url", imgURL)

			file, err := nc.downloadImage(imgURL)
			if err != nil {
				// Silently ignore files that were already downloaded.
				var fileExistsErr *fileExistsError
				if errors.As(err, &fileExistsErr) {
					return
				}
				nc.Log.Error("download error", "url", imgURL, "error", err)
				return
			}
			files = append(files, file)
		})
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		nc.Log.Error("HTTP error", "status", r.StatusCode, "url", r.Request.URL, "error", err)
	})

	c.Visit(nc.BaseDomain + "/" + nc.Account)
	nc.Log.Debug("waiting for pending requests")
	c.Wait()
	nc.Log.Info("finished scraping")

	return files, err
}

// Try to download image once.
func (nc *NitterCollector) downloadImage(url string) (string, error) {
	fileName := path.Base(url)
	dest := path.Join(nc.DownloadDir, fileName)

	// This check avoids an unnecessary copy from Colly's local cache.
	if fileExists(dest) {
		nc.Log.Debug("image already cached", "url", url)
		return dest, &fileExistsError{name: dest}
	}
	nc.Log.Info("downloading", "url", url, "dest", dest)

	c := nc.getFirefoxCollector()
	c.OnResponse(func(r *colly.Response) {
		if strings.Contains(r.Headers.Get("Content-Type"), "image") {
			fileName := path.Base(r.Request.URL.Path)
			if err := r.Save(dest); err != nil {
				nc.Log.Error("error saving file", "file", fileName, "error", err)
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
		// Hiding in plain sight...
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:142.0) Gecko/20100101 Firefox/142.0"),

		// Don't use that with a public Nitter instance.
		colly.IgnoreRobotsTxt(),

		// TODO: CacheExpiration isn't automatic in colly v2.
		colly.CacheDir("cache"),
	)
	c.OnRequest(func(r *colly.Request) {
		nc.Log.Debug("GET", "url", r.URL)
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
