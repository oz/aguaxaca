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
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

// Component is a part of the app that is run async: like the web server,
// or the cron scheduler.
type Component interface {
	// Run starts component, and blocks.
	Run(context.Context) error

	// Shutdown component, gracefully if possible.
	Shutdown(context.Context) error
}

// ShutdownGracePeriod allows 10 seconds for graceful shutdowns.
const ShutdownGracePeriod = 10

//go:embed sql/schema.sql
var ddl string

type App struct {
	DB         *sql.DB
	Ctx        context.Context
	Logger     *slog.Logger
	ListenAddr string
	Debug      bool
}

// NewApp builds the core App type.
func NewApp(ctx context.Context) *App {
	app := new(App)
	app.Ctx = ctx
	app.Logger = slog.Default()
	app.Debug = false

	return app
}

// Init starts the app: connect DB handles, etc.
func (app *App) Init(debug bool, listenAddr string) error {
	app.ListenAddr = listenAddr
	app.Debug = debug

	// debug mode: use a new logger with lower level.
	if debug {
		opts := &slog.HandlerOptions{Level: slog.LevelDebug}
		app.Logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return app.InitDB()
}

func (app *App) InitDB() error {
	// TODO: configurable path to SQLite DB
	db, err := sql.Open("sqlite", "agua.db")
	if err != nil {
		return fmt.Errorf("opening DB failed: %v", err)
	}
	app.DB = db

	// Run schema.sql: create tables, indexes, etc.
	if _, err := app.DB.ExecContext(app.Ctx, ddl); err != nil {
		return fmt.Errorf("syncing schema failed: %v", err)
	}
	return nil
}

// Start all the app's components (web server, cron scheduler, ...)
func (app *App) Start(components ...Component) error {
	// OS Signals for shutdown.
	ctx, cancel := context.WithCancel(app.Ctx)
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run components.
	errChan := make(chan error, len(components))
	var wg sync.WaitGroup
	for _, component := range components {
		if component == nil {
			return fmt.Errorf("Cannot start a nil component")
		}
		wg.Add(1)
		go func(comp Component) {
			defer wg.Done()
			if err := comp.Run(ctx); err != nil {
				app.Logger.Error("runner", "error", err)
				errChan <- err
				// Shutdown everything on error.
				cancel()
			}
		}(component)
	}

	select {
	case sig := <-sigChan:
		app.Logger.Info("shutdown", "signal", sig)
	case err := <-errChan:
		app.Logger.Error("shutdown", "error", err)
	}

	// Shutdown components.
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(app.Ctx, ShutdownGracePeriod*time.Second)
	defer shutdownCancel()

	var shutdownWg sync.WaitGroup
	shutdownErrors := make([]error, len(components))
	for _, component := range components {
		shutdownWg.Add(1)
		go func(comp Component) {
			defer shutdownWg.Done()
			if err := comp.Shutdown(shutdownCtx); err != nil {
				app.Logger.Error("shutdown", "error", err)
			}
		}(component)
	}

	// Wait for all components to stop, or timeout.
	shutdownDone := make(chan struct{})
	go func() {
		shutdownWg.Wait()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		app.Logger.Info("bye")
	case <-shutdownCtx.Done():
		app.Logger.Error("shutdown timeout")
		return fmt.Errorf("shutdown timeout exceeded")
	}

	for _, err := range shutdownErrors {
		if err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
	}

	return nil
}
