// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/clivern/finn/api"
	"github.com/clivern/finn/middleware"
	"github.com/clivern/finn/service/runner"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// InitRunner opens the local runner database and ensures the runner identity
func InitRunner() (*runner.Bolt, error) {
	path := viper.GetString("app.database.path")

	if path == "" {
		return nil, fmt.Errorf("runner init: app.database.path is required")
	}

	db := &runner.Bolt{Path: path}
	if err := db.Open(); err != nil {
		return nil, fmt.Errorf("runner init: %w", err)
	}

	mod := &runner.Module{Bolt: db, Key: &runner.Key{}}
	identity, err := mod.GetIdentity()
	if err != nil || identity == nil {
		identity, err = mod.Init()
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("runner init: %w", err)
		}
		log.Info().Str("runner_id", identity.ID).Msg("Runner initialized")
	} else {
		log.Info().Str("runner_id", identity.ID).Msg("Runner identity loaded")
	}

	fmt.Fprintln(os.Stdout, "Token:")
	fmt.Fprintln(os.Stdout, identity.ID)
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Public Key:")
	fmt.Fprintln(os.Stdout, string(identity.PublicKey))

	return db, nil
}

// SetupRunner creates a HTTP server with the health and ping endpoints.
func SetupRunner(db *runner.Bolt) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	if viper.GetInt("app.timeout") > 0 {
		timeout := time.Duration(viper.GetInt("app.timeout")) * time.Second
		r.Use(chimiddleware.Timeout(timeout))
	}
	r.Use(middleware.PrometheusMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestSizeLimit(int64(10 * 1024 * 1024)))

	deps, err := runner.NewDeps(db)
	if err != nil {
		log.Error().Err(err).Msg("Runner deps init failed")
	}

	r.Get("/_health", api.HealthAction)
	r.With(middleware.BasicAuth(
		viper.GetString("app.metrics.username"),
		viper.GetString("app.metrics.secret"),
	)).Get("/_metrics", promhttp.Handler().ServeHTTP)
	r.With(middleware.RunnerToken(deps)).Post("/_ping", api.NewPingHandler(deps))

	return r
}

// RunRunner starts the runner HTTP server with graceful shutdown.
func RunRunner(handler http.Handler, db *runner.Bolt) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", strconv.Itoa(viper.GetInt("app.port"))),
		Handler: handler,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info().
			Int("port", viper.GetInt("app.port")).
			Msg("Starting runner")

		var err error
		if viper.GetBool("app.tls.status") {
			err = srv.ListenAndServeTLS(
				viper.GetString("app.tls.crt_path"),
				viper.GetString("app.tls.key_path"),
			)
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("runner error: %w", err)
	case sig := <-quit:
		log.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal")

		shutdownTimeout := 30 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		log.Info().
			Dur("timeout", shutdownTimeout).
			Msg("Gracefully shutting down runner")

		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("runner forced to shutdown: %w", err)
		}

		if db != nil {
			if err := db.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close runner database")
			}
		}

		log.Info().Msg("Runner shutdown complete")
	}

	return nil
}
