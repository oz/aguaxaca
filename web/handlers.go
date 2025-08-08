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
	"net/http"
	"time"

	"git.cypr.io/oz/aguaxaca/app/db"
)

func (s *Server) RootHandler(w http.ResponseWriter, r *http.Request) {
	// Create a queries object with the database connection
	queries := db.New(s.app.DB)

	// Fetch recent deliveries from the database
	deliveries, err := queries.ListDeliveries(r.Context(), db.UnixTime{Time: startAt()})
	if err != nil {
		s.app.Logger.Error("failed to list deliveries", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Pass deliveries to the template
	err = s.templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Deliveries": deliveries,
	})
	if err != nil {
		s.app.Logger.Error("failed to render template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Compute the zone offset just once.
var utcMinus6 = time.FixedZone("UTC-6", -6*60*60)

// 7 days ago, in UTC-6.
func startAt() time.Time {
	now := time.Now().In(utcMinus6)
	return now.AddDate(0, 0, -7)
}
