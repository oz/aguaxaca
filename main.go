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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.cypr.io/oz/aguaxaca/app"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {
	ctx := context.Background()
	app := app.NewApp(ctx)

	// CLI command: aguaxaca collect
	collectCmd := &ffcli.Command{
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

	// CLI command: aguaxaca analyze
	analyzeCmd := &ffcli.Command{
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

	// CLI command: aguaxaca server
	serverCmd := &ffcli.Command{
		Name:      "server",
		ShortHelp: "Start web server",
		Exec: func(context.Context, []string) error {
			fmt.Println("start web server")
			return nil
		},
	}

	// root command
	rootFlagSet := flag.NewFlagSet("aguaxaca", flag.ExitOnError)
	debug := rootFlagSet.Bool("debug", false, "log debug information")
	listenAddr = fs.String("listen", "localhost:8080", "listen address")
	root := &ffcli.Command{
		Name:        "aguaxaca",
		ShortUsage:  "aguaxaca [OPTIONS] SUBCOMMAND ...",
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{collectCmd, analyzeCmd, serverCmd},
		Exec: func(context.Context, []string) error {
			// The root command by itself has no use. Show usage help.
			return flag.ErrHelp
		},
	}

	if err := root.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Configure default *App after flags parsing...
	if err := app.Init(*debug); err != nil {
		fmt.Fprintf(os.Stderr, "App init error: %v\n", err)
		os.Exit(2)
	}

	if err := root.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
