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

	_ "modernc.org/sqlite"
)

//go:embed sql/schema.sql
var ddl string

type App struct {
	DB     *sql.DB
	Ctx    context.Context
	Logger *slog.Logger
}

// NewApp builds the core App type.
func NewApp(ctx context.Context) *App {
	app := new(App)
	app.Ctx = ctx
	app.Logger = slog.Default()

	return app
}

// Init starts the app: connect DB handles, etc.
func (app *App) Init(debug bool) error {
	// Create a new text logger for debug-mode
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
