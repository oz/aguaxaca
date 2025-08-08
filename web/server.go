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

package web

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"

	"git.cypr.io/oz/aguaxaca/app"
)

// RequestTimeOut is 60 seconds
const RequestTimeOut = 60

// ShutdownGracePeriod allows 10 seconds for graceful shutdown.
const ShutdownGracePeriod = 10

type Server struct {
	app       *app.App
	templates *template.Template
}

func NewServer(app *app.App) *Server {
	tmpl := template.Must(
		template.ParseFiles(
			"web/templates/layout.html",
			"web/templates/index.html",
		),
	)
	return &Server{
		app:       app,
		templates: tmpl,
	}
}

// NewHandler configure server routes and middlewares.
func (s *Server) NewHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(s.app.Logger, s.loggerOptions()))
	r.Use(middleware.Recoverer)

	// Timeout requests after 60s
	r.Use(middleware.Timeout(RequestTimeOut * time.Second))

	// Routes
	r.Get("/", s.RootHandler)

	return r
}

// Run starts an http.Server
func (s *Server) Run() error {
	s.app.Logger.Info("starting web server", "address", s.app.ListenAddr)
	server := &http.Server{Addr: s.app.ListenAddr, Handler: s.NewHandler()}

	// Ctrl-c ...
	serverCtx, serverStopCtx := context.WithCancel(s.app.Ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig

		// Allow 10s for graceful shutdown.
		s.app.Logger.Info("shutting down")
		shutdownCtx, _ := context.WithTimeout(serverCtx, ShutdownGracePeriod*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				s.app.Logger.Error("graceful shutdown timed out")
			}
		}()

		// Graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			s.app.Logger.Error("shutdown", "error", err)
			os.Exit(2)
		}
		serverStopCtx()
	}()

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		s.app.Logger.Error("ListenAndServe", "error", err)
		os.Exit(2)
	}
	<-serverCtx.Done()

	return nil
}

func (s *Server) loggerOptions() *httplog.Options {
	opts := httplog.Options{
		// Level defines the verbosity of the request logs:
		// slog.LevelDebug - log all responses (incl. OPTIONS)
		// slog.LevelInfo  - log responses (excl. OPTIONS)
		// slog.LevelWarn  - log 4xx and 5xx responses only (except for 429)
		// slog.LevelError - log 5xx responses only
		Level: slog.LevelInfo,

		// Set log output to Elastic Common Schema (ECS) format.
		Schema: httplog.SchemaECS,

		// RecoverPanics recovers from panics occurring in the underlying HTTP handlers
		// and middlewares. It returns HTTP 500 unless response status was already set.
		//
		// NOTE: Panics are logged as errors automatically, regardless of this setting.
		RecoverPanics: true,

		// Optionally, filter out some request logs.
		// Skip: func(req *http.Request, respStatus int) bool {
		// 	return respStatus == 404 || respStatus == 405
		// },

		// Optionally, log selected request/response headers explicitly.
		// LogRequestHeaders:  []string{"Origin"},
		// LogResponseHeaders: []string{},

		LogRequestBody:  isRequestLoggingEnabled(s.app.Debug),
		LogResponseBody: isRequestLoggingEnabled(s.app.Debug),
	}
	if s.app.Debug {
		opts.Level = slog.LevelDebug
	}
	return &opts
}

func isRequestLoggingEnabled(debug bool) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		return debug
	}
}
