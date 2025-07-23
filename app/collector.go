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
	"hash/fnv"
	"log"
	"os"

	"git.cypr.io/oz/aguaxaca/app/db"
	"git.cypr.io/oz/aguaxaca/collector"
)

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
	// Download images
	images, err := c.collector.DownloadImages()
	if err != nil {
		return fmt.Errorf("failed image download: %v", err)
	}
	log.Println("DEBUG: found images:", images)

	// Create import jobs for each new image.
	for _, path := range images {
		fileHash, err := hashFile(path)
		if err != nil {
			log.Printf("Error hashing file %s: %v", path, err)
			continue
		}

		if err := c.CreateImportIfNotExists(path, fileHash); err != nil {
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

	// Insert new import
	imp, err := queries.CreateImport(c.app.Ctx, db.CreateImportParams{FilePath: path, FileHash: hash})
	if err != nil {
		return fmt.Errorf("create import record for '%s': %v", path, err)
	}
	log.Printf("Created new import job: #%d", imp.ID)
	return nil

}

func hashFile(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	h := fnv.New64a()
	h.Write(data)
	u := h.Sum64()
	return int64(u), nil
}
