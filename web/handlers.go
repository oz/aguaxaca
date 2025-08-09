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

package web

import (
	"database/sql"
	"html"
	"net/http"
	"strings"

	"git.cypr.io/oz/aguaxaca/app/db"
)

func (s *Server) RootHandler(w http.ResponseWriter, r *http.Request) {
	nameParam := r.URL.Query().Get("name")
	deliveries, err := findDeliveries(r, s.app.DB, queryParamToFTS(nameParam))
	if err != nil {
		s.app.Logger.Error("failed to list deliveries", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render HTML.
	err = s.templates.ExecuteTemplate(w, "index.html", map[string]any{
		"Deliveries": deliveries,
		"Name":       html.EscapeString(nameParam),
	})
	if err != nil {
		s.app.Logger.Error("failed to render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// findDeliveries for the home: either the latest, or FTS on name.
func findDeliveries(r *http.Request, conn *sql.DB, nameSearch string) ([]db.Delivery, error) {
	queries := db.New(conn)
	if nameSearch == "" {
		fromDate := db.UnixTime{Time: daysAgo(7)}
		return queries.ListDeliveries(r.Context(), fromDate)
	}

	params := db.SearchDeliveriesByNameParams{
		LocationName: nameSearch,
		Date:         db.UnixTime{Time: daysAgo(90)},
	}
	return queries.SearchDeliveriesByName(r.Context(), params)
}

// Basic search query param cleanup.
func queryParamToFTS(param string) string {
	trimmed := strings.ToLower(strings.TrimSpace(param))
	if trimmed == "" {
		return ""
	}

	// Quote some FTS5 search tokens that break too easily.
	// See https://www.sqlite.org/fts5.html
	r := strings.NewReplacer(
		`"`, `""`,
		"(", `"("`,
		")", `")"`,
		",", "",
	)
	trimmed = r.Replace(trimmed)

	// TODO: remove wildcard matches, like "prefix*": they will become too
	//       slow as the DB grows.
	return trimmed
}
