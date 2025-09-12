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
	"context"

	"git.cypr.io/oz/aguaxaca/app"
	"github.com/go-co-op/gocron/v2"
)

type Workers struct {
	app   *app.App
	sched gocron.Scheduler
}

func NewScheduler(app *app.App) *Workers {
	sched, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}
	return &Workers{app, sched}
}

// Run go-cron scheduler until we're told to stop.
func (w *Workers) Run(ctx context.Context) error {
	w.app.Logger.Info("starting cron scheduler")
	w.sched.Start()
	<-ctx.Done()
	return nil
}

func (w *Workers) Shutdown(_ context.Context) error {
	w.app.Logger.Info("shutting down cron scheduler")
	return w.sched.Shutdown()
}
