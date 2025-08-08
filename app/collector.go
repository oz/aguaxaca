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

package app

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dchest/siphash"

	"git.cypr.io/oz/aguaxaca/app/db"
	"git.cypr.io/oz/aguaxaca/collector"
)

// SipHashKey is a prefectly random key (used to dedup files, not sensitive).
var SipHashKey = [16]byte{
	0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF,
	0xFE, 0xDC, 0xBA, 0x98, 0x76, 0x54, 0x32, 0x10,
}

type Collector struct {
	collector collector.Collector
	app       *App
}

func (app *App) NewCollector(collector collector.Collector) *Collector {
	return &Collector{
		app:       app,
		collector: collector,
	}
}

// Collect runs an image collector to fetch images, and create import
// records in the local DB.
func (c *Collector) Collect() error {
	images, err := c.collector.DownloadImages()
	if err != nil {
		return fmt.Errorf("image download: %v", err)
	}
	log.Println("DEBUG: found images:", images)

	// Create import jobs for each new image.
	for _, path := range images {
		fileHash, err := hashFile(path)
		if err != nil {
			log.Printf("Error hashing file %s: %v", path, err)
			continue
		}

		if err := c.CreateImportIfNotExists(path, int64(fileHash)); err != nil {
			log.Printf("Error importing %s: %v", path, err)
			continue
		}
	}
	return nil
}

func (c *Collector) CreateImportIfNotExists(path string, hash int64) error {
	queries := db.New(c.app.DB)
	count, err := queries.CountImportsByHash(c.app.Ctx, hash)
	if err != nil {
		return fmt.Errorf("lookup import record for '%s': %v", path, err)
	}
	if count != 0 {
		log.Printf("skipping '%s': already imported (hash: %d)\n", path, hash)
		return nil
	}

	imp, err := queries.CreateImport(c.app.Ctx, db.CreateImportParams{FilePath: path, FileHash: hash})
	if err != nil {
		return fmt.Errorf("create import record for '%s': %v", path, err)
	}
	log.Printf("Created new import job #%d for %s", imp.ID, imp.FilePath)

	return nil
}

// hashFile uses siphash for deduplication.
func hashFile(filepath string) (uint64, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	hasher := siphash.New(SipHashKey[:])
	if _, err := io.Copy(hasher, file); err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}
