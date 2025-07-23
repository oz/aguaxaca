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
	"log"

	"git.cypr.io/oz/aguaxaca/app/db"
)

type Analyzer struct {
	app *App
}

func (app *App) NewAnalyzer() *Analyzer {
	return &Analyzer{app: app}
}

func (analyzer *Analyzer) ProcessPendingImports() (int, error) {
	queries := db.New(analyzer.app.DB)
	imports, err := queries.GetPendingImports(analyzer.app.Ctx)
	if err != nil {
		return 0, err
	}
	imCount := 0
	for _, im := range imports {
		log.Printf("Analyzing import #%d: %s", im.ID, im.FilePath)
	}
	return imCount, fmt.Errorf("not implemented")
}
