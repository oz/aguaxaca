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
	"flag"

	"git.cypr.io/oz/aguaxaca/app"
	"git.cypr.io/oz/aguaxaca/web"
	"git.cypr.io/oz/aguaxaca/workers"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func ServerCommand(app *app.App) *ffcli.Command {
	fs := flag.NewFlagSet("aguaxaca server", flag.ExitOnError)
	listenAddr := fs.String("listen", "localhost:8080", "listen address")

	return &ffcli.Command{
		Name:      "server",
		ShortHelp: "Start web server + async workers",
		FlagSet:   fs,
		Exec: func(context.Context, []string) error {
			app.ListenAddr = *listenAddr
			serv := web.NewServer(app)
			sched := workers.NewScheduler(app)
			return app.Start(serv, sched)
		},
	}
}
