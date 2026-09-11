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

package cmd

import (
	"context"
	"fmt"

	"git.cypr.io/oz/aguaxaca/app"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func AnalyzeCommand(app *app.App) *ffcli.Command {
	return &ffcli.Command{
		Name:      "analyze",
		ShortHelp: "Analyze and extract data from collected images",
		Exec: func(context.Context, []string) error {
			analyzer := app.NewAnalyzer()
			count, err := analyzer.ProcessPendingImports()
			if err != nil {
				fmt.Printf("Error analyzing images: %v", err)
			}
			fmt.Printf("Image analysis complete (%d).\n", count)
			return nil
		},
	}
}
