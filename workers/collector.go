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

package workers

import (
	"database/sql"
	"log/slog"
	"time"

	"git.cypr.io/oz/aguaxaca/app"
	"git.cypr.io/oz/aguaxaca/app/db"
)

// CollectorGracePeriod is the number of hours we wait after the last
// successful import, before checking for new reports.
const CollectorGracePeriod = 12

// An hourly job to run app.DefaultCollector.
func collectorJob(app *app.App) error {
	log := app.Logger.With("job", "collector")
	if !shouldRunCollector(app, log) {
		return nil
	}

	// Run collector.
	if err := app.DefaultCollector().Collect(); err != nil {
		log.Error("Collect", "error", err)
		return err
	}

	// Parse new reports.
	analyzer := app.NewAnalyzer()
	count, err := analyzer.ProcessPendingImports()
	if err != nil {
		log.Error("Collect", "ProcessPendingImports", err)
		return err
	}
	log.Info("Analyze complete", "count", count)

	return nil
}

// If the latest import was completed less than CollectorGracePeriod
// (12) hours ago, we don't need to check for a new report yet.
// After that, run on the hour to look for new data.
func shouldRunCollector(app *app.App, log *slog.Logger) bool {
	queries := db.New(app.DB)
	latest, err := queries.GetLatestImport(app.Ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return true
		}
		log.Error("GetLatestImport", "error", err)
		return false
	}

	nextRunTime := latest.CreatedAt.Time.Add(time.Duration(CollectorGracePeriod * time.Hour))
	now := time.Now().UTC()
	return nextRunTime.Before(now)
}
