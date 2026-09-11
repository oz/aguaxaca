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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.cypr.io/oz/aguaxaca/app"
	"git.cypr.io/oz/aguaxaca/cmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {
	ctx := context.Background()
	app := app.NewApp(ctx)

	// Flags & sub-commands
	fs := flag.NewFlagSet("aguaxaca", flag.ExitOnError)
	debug := fs.Bool("debug", false, "log debug information")
	//listenAddr := fs.String("listen", "localhost:8080", "listen address")
	root := &ffcli.Command{
		Name:       "aguaxaca",
		ShortUsage: "aguaxaca [OPTIONS] SUBCOMMAND ...",
		FlagSet:    fs,
		Subcommands: []*ffcli.Command{
			cmd.CollectCommand(app),
			cmd.AnalyzeCommand(app),
			cmd.ServerCommand(app),
		},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}
	if err := root.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Configure App based on flags.
	if err := app.Init(*debug); err != nil {
		fmt.Fprintf(os.Stderr, "App init error: %v\n", err)
		os.Exit(2)
	}

	if err := root.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
