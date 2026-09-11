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
	"os"

	"git.cypr.io/oz/aguaxaca/app"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// CLI command: aguaxaca collect
func CollectCommand(app *app.App) *ffcli.Command {
	return &ffcli.Command{
		Name:      "collect",
		ShortHelp: "Fetch latest water schedules",
		Exec: func(context.Context, []string) error {
			if err := app.DefaultCollector().Collect(); err != nil {
				fmt.Printf("Error collecting schedules: %v\n", err)
				os.Exit(2)
			}

			return nil
		},
	}
}
