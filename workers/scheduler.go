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
	"time"

	"git.cypr.io/oz/aguaxaca/app"
	"github.com/go-co-op/gocron/v2"
)

type Scheduler struct {
	app   *app.App
	sched gocron.Scheduler
}

func NewScheduler(app *app.App) *Scheduler {
	sched, err := gocron.NewScheduler(gocron.WithLogger(app.Logger))
	if err != nil {
		app.Logger.Error("scheduler", "error", err)
		return nil
	}

	// Run collectorJob hourly.
	if _, err = sched.NewJob(
		gocron.DurationJob(1*time.Hour),
		gocron.NewTask(collectorJob, app),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	); err != nil {
		app.Logger.Error("scheduler", "error", err)
		return nil
	}

	return &Scheduler{app, sched}
}

// Run go-cron scheduler until we're told to stop.
func (s *Scheduler) Run(ctx context.Context) error {
	s.app.Logger.Info("starting cron scheduler")
	s.sched.Start()
	<-ctx.Done()
	return nil
}

func (s *Scheduler) Shutdown(_ context.Context) error {
	s.app.Logger.Info("shutting down cron scheduler")
	return s.sched.Shutdown()
}
