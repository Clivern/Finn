// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package core

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/clivern/finn/api"
	"github.com/clivern/finn/db"
	"github.com/clivern/finn/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Setup creates and configures the HTTP server
func SetupServer(Static embed.FS) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	if viper.GetInt("app.timeout") > 0 {
		timeout := time.Duration(viper.GetInt("app.timeout")) * time.Second
		r.Use(chimiddleware.Timeout(timeout))
	}
	r.Use(middleware.PrometheusMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestSizeLimit(int64(10 * 1024 * 1024)))
	r.Use(middleware.SessionAuth())
	r.Get("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r.Group(func(r chi.Router) {
		r.Get("/api/v1/public/_health", api.HealthAction)
		r.Get("/api/v1/public/_ready", api.ReadyAction)
		r.Post("/api/v1/public/action/setup", api.SetupAction)
		r.Get("/api/v1/public/action/setup/status", api.SetupStatusAction)
		r.Post("/api/v1/public/action/login", api.LoginAction)
		r.Post("/api/v1/public/action/logout", api.LogoutAction)
		r.Post("/api/v1/public/action/forgot-password", api.ForgotPasswordAction)
		r.Post("/api/v1/public/action/reset-password", api.ResetPasswordAction)
		r.Get("/api/v1/public/action/invite/token/{token}", api.GetUserInviteByTokenAction)
		r.Post("/api/v1/public/action/register", api.RegisterAction)
		r.Post("/client/{slug}/state", api.CreateBackendStateAction)
		r.Get("/client/{slug}/state", api.GetBackendStateAction)
		r.Post("/client/{slug}/lock", api.LockBackendStateAction)
		r.Post("/client/{slug}/unlock", api.UnlockBackendStateAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin, db.UserRoleRegular))
		r.Get("/api/v1/action/profile", api.GetProfileAction)
		r.Put("/api/v1/action/profile", api.UpdateProfileAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin))
		r.Put("/api/v1/action/settings", api.UpdateSettingsAction)
		r.Get("/api/v1/action/settings", api.GetSettingsAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin, db.UserRoleRegular))
		r.Post("/api/v1/apiKeys", api.CreateUserAPIKeyAction)
		r.Get("/api/v1/apiKeys", api.ListUserAPIKeysAction)
		r.Get("/api/v1/apiKeys/{id}", api.GetUserAPIKeyAction)
		r.Delete("/api/v1/apiKeys/{id}", api.DeleteUserAPIKeyAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin))
		r.Post("/api/v1/users", api.CreateUserAction)
		r.Get("/api/v1/users", api.ListUsersAction)
		r.Get("/api/v1/users/{id}", api.GetUserAction)
		r.Put("/api/v1/users/{id}", api.UpdateUserAction)
		r.Delete("/api/v1/users/{id}", api.DeleteUserAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin))
		r.Post("/api/v1/invites", api.CreateUserInviteAction)
		r.Get("/api/v1/invites", api.ListInvitesAction)
		r.Get("/api/v1/invites/{id}", api.GetUserInviteAction)
		r.Delete("/api/v1/invites/{id}", api.DeleteUserInviteAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin))
		r.Get("/api/v1/runners", api.ListRunnersAction)
		r.Get("/api/v1/runners/{id}", api.GetRunnerAction)
		r.Post("/api/v1/runners", api.CreateRunnerAction)
		r.Put("/api/v1/runners/{id}", api.UpdateRunnerAction)
		r.Delete("/api/v1/runners/{id}", api.DeleteRunnerAction)
		r.Post("/api/v1/runners/{id}/ping", api.PingRunnerAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin, db.UserRoleRegular))
		r.Post("/api/v1/workspaces", api.CreateWorkspaceAction)
		r.Get("/api/v1/workspaces", api.ListWorkspacesAction)
		r.Get("/api/v1/workspaces/{id}", api.GetWorkspaceAction)
		r.Put("/api/v1/workspaces/{id}", api.UpdateWorkspaceAction)
		r.Delete("/api/v1/workspaces/{id}", api.DeleteWorkspaceAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin, db.UserRoleRegular))
		r.Post("/api/v1/workspaces/{workspaceID}/users", api.AddWorkspaceUserAction)
		r.Get("/api/v1/workspaces/{workspaceID}/users", api.ListWorkspaceUsersAction)
		r.Get("/api/v1/workspaces/{workspaceID}/all-users", api.ListAllUsersAction)
		r.Get("/api/v1/workspaces/{workspaceID}/users/{userID}", api.GetWorkspaceUserAction)
		r.Put("/api/v1/workspaces/{workspaceID}/users/{userID}", api.UpdateWorkspaceUserRoleAction)
		r.Delete("/api/v1/workspaces/{workspaceID}/users/{userID}", api.RemoveWorkspaceUserAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin, db.UserRoleRegular))
		r.Post("/api/v1/workspaces/{workspaceID}/repositories", api.CreateWorkspaceRepositoryAction)
		r.Get("/api/v1/workspaces/{workspaceID}/repositories", api.ListWorkspaceRepositoriesAction)
		r.Get("/api/v1/workspaces/{workspaceID}/repositories/{id}", api.GetWorkspaceRepositoryAction)
		r.Put("/api/v1/workspaces/{workspaceID}/repositories/{id}", api.UpdateWorkspaceRepositoryAction)
		r.Delete("/api/v1/workspaces/{workspaceID}/repositories/{id}", api.DeleteWorkspaceRepositoryAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin, db.UserRoleRegular))
		r.Post("/api/v1/workspaces/{workspaceID}/keys", api.CreateWorkspaceSSHKeyAction)
		r.Get("/api/v1/workspaces/{workspaceID}/keys", api.ListWorkspaceSSHKeysAction)
		r.Get("/api/v1/workspaces/{workspaceID}/keys/{id}", api.GetWorkspaceSSHKeyAction)
		r.Put("/api/v1/workspaces/{workspaceID}/keys/{id}", api.UpdateWorkspaceSSHKeyAction)
		r.Delete("/api/v1/workspaces/{workspaceID}/keys/{id}", api.DeleteWorkspaceSSHKeyAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin, db.UserRoleRegular))
		r.Post("/api/v1/workspaces/{workspaceID}/credentials", api.CreateWorkspaceGitCredAction)
		r.Get("/api/v1/workspaces/{workspaceID}/credentials", api.ListWorkspaceGitCredsAction)
		r.Get("/api/v1/workspaces/{workspaceID}/credentials/{id}", api.GetWorkspaceGitCredAction)
		r.Put("/api/v1/workspaces/{workspaceID}/credentials/{id}", api.UpdateWorkspaceGitCredAction)
		r.Delete("/api/v1/workspaces/{workspaceID}/credentials/{id}", api.DeleteWorkspaceGitCredAction)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(db.UserRoleAdmin, db.UserRoleRegular))
		r.Post("/api/v1/workspaces/{workspaceID}/backends", api.CreateWorkspaceBackendAction)
		r.Get("/api/v1/workspaces/{workspaceID}/backends", api.ListWorkspaceBackendsAction)
		r.Get("/api/v1/workspaces/{workspaceID}/backends/{id}", api.GetWorkspaceBackendAction)
		r.Get("/api/v1/workspaces/{workspaceID}/backends/{id}/password", api.RevealWorkspaceBackendPasswordAction)
		r.Put("/api/v1/workspaces/{workspaceID}/backends/{id}", api.UpdateWorkspaceBackendAction)
		r.Delete("/api/v1/workspaces/{workspaceID}/backends/{id}", api.DeleteWorkspaceBackendAction)
	})

	r.With(middleware.BasicAuth(
		viper.GetString("app.metrics.username"),
		viper.GetString("app.metrics.secret"),
	)).Get("/api/v1/public/_metrics", promhttp.Handler().ServeHTTP)

	dist, err := fs.Sub(Static, "web/dist")
	if err != nil {
		panic(fmt.Sprintf(
			"Error while accessing dist files: %s",
			err.Error(),
		))
	}

	r.Handle("/assets/*", http.StripPrefix("/", http.FileServer(http.FS(dist))))
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		indexFile, err := dist.Open("index.html")
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile.(io.ReadSeeker))
	})

	return r
}

// InitDatabase initializes the database connection from configuration
func InitDatabase() error {
	dbConfig := db.Config{
		Driver:          viper.GetString("app.database.driver"),
		Host:            viper.GetString("app.database.host"),
		Port:            viper.GetInt("app.database.port"),
		Username:        viper.GetString("app.database.username"),
		Password:        viper.GetString("app.database.password"),
		Database:        viper.GetString("app.database.name"),
		MaxOpenConns:    viper.GetInt("app.database.max_open_conns"),
		MaxIdleConns:    viper.GetInt("app.database.max_idle_conns"),
		ConnMaxLifetime: viper.GetInt("app.database.conn_max_lifetime"),
		DataSource:      viper.GetString("app.database.datasource"),
	}

	return db.InitDB(dbConfig)
}

// Run starts the HTTP server with graceful shutdown support
func RunServer(handler http.Handler) error {
	if err := InitDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Error().Err(err).Msg("Error closing database connection")
		}
	}()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", strconv.Itoa(viper.GetInt("app.port"))),
		Handler: handler,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info().
			Int("port", viper.GetInt("app.port")).
			Bool("tls", viper.GetBool("app.tls.status")).
			Msg("Starting HTTP server")

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
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		log.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal")

		shutdownTimeout := 30 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		log.Info().
			Dur("timeout", shutdownTimeout).
			Msg("Gracefully shutting down server")

		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}

		log.Info().Msg("Server shutdown complete")
	}

	return nil
}
