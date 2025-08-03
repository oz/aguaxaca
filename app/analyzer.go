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
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"git.cypr.io/oz/aguaxaca/app/db"
	"git.cypr.io/oz/aguaxaca/parser"
)

// MaxRuns limits how many retries the analyzer performs on errors.
const MaxRuns = 3

// DateFormat is the format of date fields in the LLM's output.
const DateFormat = "2006-01-02"

type Analyzer struct {
	app *App
}

func (app *App) NewAnalyzer() *Analyzer {
	return &Analyzer{app: app}
}

func (analyzer *Analyzer) ProcessPendingImports() (int, error) {
	queries := db.New(analyzer.app.DB)

	runs := sql.NullInt64{Int64: MaxRuns, Valid: true}
	imports, err := queries.GetPendingImports(analyzer.app.Ctx, runs)
	if err != nil {
		return 0, err
	}

	imCount := 0
	for _, im := range imports {
		log.Printf("Analyzing import #%d: %s", im.ID, im.FilePath)

		// Query Anthropic (seq)
		csvData, err := parser.ParseFileWithPrompt(analyzer.app.Ctx, im.FilePath, parser.DefaultCSVPrompt)
		if err != nil {
			log.Printf("Parser error for #%d: %v", im.ID, err)

			if dbErr := queries.FailImport(analyzer.app.Ctx, im.ID); dbErr != nil {
				return imCount, fmt.Errorf("Error updating DB (FailImport) for #%d: %v", im.ID, dbErr)
			}
			continue
		}
		log.Printf("Parsed image #%d", im.ID)

		// Import csv data
		if err := analyzer.ImportData(&im, csvData); err != nil {
			log.Printf("Parser error for #%d: %v", im.ID, err)

			if dbErr := queries.FailImport(analyzer.app.Ctx, im.ID); dbErr != nil {
				return imCount, fmt.Errorf("Error updating DB (FailImport) for #%d: %v", im.ID, dbErr)
			}
			continue
		}

		// Update import state
		if err := queries.CompleteImport(analyzer.app.Ctx, im.ID); err != nil {
			return imCount, fmt.Errorf("Error updating DB (CompleteImport) for #%d: %v", im.ID, err)
		}
		imCount += 1
	}

	return imCount, nil
}

func (analyzer *Analyzer) ImportData(im *db.Import, csvData string) error {
	log.Printf("DEBUG CSV data for #%d:\n%s\n\n", im.ID, csvData)

	queries := db.New(analyzer.app.DB)
	reader := csv.NewReader(strings.NewReader(csvData))

	// Read header row
	_, err := reader.Read()
	if err != nil {
		return fmt.Errorf("error reading CSV header: %w", err)
	}

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading CSV record: %w", err)
		}

		date, err := time.Parse(DateFormat, record[0])
		if err != nil {
			return fmt.Errorf("invalid date format '%s': %w", record[0], err)
		}

		// Create delivery record with lowercase location_type
		_, err = queries.CreateDelivery(analyzer.app.Ctx, db.CreateDeliveryParams{
			Date:         db.UnixTime{Time: date},
			Schedule:     strings.ToLower(record[1]),
			LocationType: strings.ToLower(record[2]), // Ensure lowercase
			LocationName: record[3],
		})
		if err != nil {
			return fmt.Errorf("failed to create delivery: %w", err)
		}
	}

	return nil
}
