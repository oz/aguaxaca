// This file is part of Aguaxaca.
// Copyright (C) 2026 Arnaud Berthomier.
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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const TwitterAPIBaseURL = "https://api.twitterapi.io"

type TwitterAPICollector struct {
	Account     string
	APIKey      string
	DownloadDir string
	BaseURL     string
	Log         *slog.Logger
	HTTPClient  *http.Client
}

func NewTwitterAPICollector(account, apiKey string) *TwitterAPICollector {
	return &TwitterAPICollector{
		Account:     account,
		APIKey:      apiKey,
		DownloadDir: DefaultDownloadDir,
		BaseURL:     TwitterAPIBaseURL,
		Log:         slog.Default(),
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

type twitterAPITweet struct {
	Text             string `json:"text"`
	ExtendedEntities struct {
		Media []struct {
			Type          string `json:"type"`
			MediaURL      string `json:"media_url_https"`
			MediaURLHttps string `json:"mediaUrlHttps"`
		} `json:"media"`
	} `json:"extendedEntities"`
}

type twitterAPIResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Code   int    `json:"code"`
	Data   struct {
		Tweets []twitterAPITweet `json:"tweets"`
	} `json:"data"`
}

// DownloadImages fetches the account's latest tweets (one page),
// and downloads the photos attached to the ones carrying the
// "HoyLlegaElAgua" hashtag.
func (tc *TwitterAPICollector) DownloadImages() ([]string, error) {
	files := []string{}

	if err := os.MkdirAll(tc.DownloadDir, 0750); err != nil {
		return nil, err
	}

	tweets, err := tc.fetchLatestTweets()
	if err != nil {
		return nil, err
	}

	for _, tweet := range tweets {
		if !strings.Contains(tweet.Text, "HoyLlegaElAgua") {
			continue
		}

		for _, media := range tweet.ExtendedEntities.Media {
			if media.Type != "photo" {
				continue
			}
			imgURL := media.MediaURL
			if imgURL == "" {
				imgURL = media.MediaURLHttps
			}
			if imgURL == "" {
				continue
			}
			// Request the full-resolution image.
			sep := "?"
			if strings.Contains(imgURL, "?") {
				sep = "&"
			}
			imgURL += sep + "name=orig"
			tc.Log.Info("found image", "url", imgURL)

			file, err := tc.downloadImage(imgURL)
			if err != nil {
				// Silently ignore files that were already downloaded.
				if _, ok := errors.AsType[*fileExistsError](err); ok {
					continue
				}
				tc.Log.Error("download error", "url", imgURL, "error", err)
				continue
			}
			files = append(files, file)
		}
	}

	return files, nil
}

// fetchLatestTweets fetches the first page of the account's tweets, newest
// first. Only the first page is fetched, which is all we need given the
// collection frequency.
func (tc *TwitterAPICollector) fetchLatestTweets() ([]twitterAPITweet, error) {
	u, err := url.Parse(strings.TrimSuffix(tc.BaseURL, "/") + "/twitter/user/last_tweets")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("userName", tc.Account)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", tc.APIKey)
	req.Header.Set("Accept", "application/json")

	tc.Log.Debug("GET", "url", u.String())
	resp, err := tc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("last_tweets request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("last_tweets: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed twitterAPIResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if parsed.Status != "" && parsed.Status != "success" {
		msg := parsed.Msg
		if msg == "" {
			msg = "unknown error"
		}
		return nil, fmt.Errorf("last_tweets: %s", msg)
	}

	return parsed.Data.Tweets, nil
}

// downloadImage downloads a single image, naming it after the URL path the
// same way the Nitter collector does (/media/ABC.jpg -> media%2FABC.jpg).
func (tc *TwitterAPICollector) downloadImage(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Percent-encode the path into a flat filename, like the Nitter collector.
	name := strings.TrimPrefix(u.EscapedPath(), "/")
	name = strings.ReplaceAll(name, "/", "%2F")
	dest := path.Join(tc.DownloadDir, name)

	if fileExists(dest) {
		tc.Log.Debug("image already cached", "url", rawURL)
		return dest, &fileExistsError{name: dest}
	}
	tc.Log.Info("downloading", "url", rawURL, "dest", dest)

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:160.0) Gecko/20100101 Firefox/160.0")
	req.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,*/*")

	resp, err := tc.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "image") {
		return "", fmt.Errorf("unexpected content-type %q", ct)
	}

	// Write to a temp file first, so a partial download doesn't end up
	// being treated as a cached file on the next run.
	tmp, err := os.CreateTemp(tc.DownloadDir, ".download-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return "", err
	}

	return dest, nil
}
