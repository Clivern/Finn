// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package core

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clivern/finn/pkg/runner"
	_ "github.com/clivern/finn/pkg/runner/tasks"
	runv1 "github.com/clivern/finn/proto/runner/v1"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// InitRunner opens the local runner database and ensures the runner identity
func InitRunner() (*runner.Bolt, error) {
	db := &runner.Bolt{
		Path: viper.GetString("app.database.path"),
	}
	if db.Path == "" {
		return nil, fmt.Errorf("runner init: app.database.path is required")
	}

	if err := db.Open(); err != nil {
		return nil, fmt.Errorf("runner init: %w", err)
	}

	mod := &runner.Module{Bolt: db}
	identity, err := mod.GetIdentity()

	if err != nil || identity == nil {
		identity, err = mod.Init()
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("runner init: %w", err)
		}
		log.Info().
			Str("runner_id", identity.ID).
			Msg("Runner initialized")
	} else {
		log.Info().
			Str("runner_id", identity.ID).
			Msg("Runner identity loaded")
	}

	fmt.Fprintln(os.Stdout, "ID:")
	fmt.Fprintln(os.Stdout, identity.ID)
	fmt.Fprintln(os.Stdout, "Token:")
	fmt.Fprintln(os.Stdout, identity.Token)

	return db, nil
}

// RunRunner starts the runner gRPC server with graceful shutdown.
func RunRunner(db *runner.Bolt) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deps, err := runner.NewDeps(db)
	if err != nil {
		return fmt.Errorf("runner deps: %w", err)
	}

	log.Info().Msg("Starting runner worker")
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		deps.Worker.Run(ctx, deps)
	}()

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			runner.LoggingUnaryInterceptor(),
			runner.AuthUnaryInterceptor(deps),
		),
	}
	if viper.GetBool("app.tls.status") {
		creds, err := credentials.NewServerTLSFromFile(
			viper.GetString("app.tls.crt_path"),
			viper.GetString("app.tls.key_path"),
		)
		if err != nil {
			return fmt.Errorf("runner gRPC TLS: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	srv := grpc.NewServer(opts...)
	runv1.RegisterRunnerServiceServer(srv, runner.NewRunnerGRPCService(deps))
	reflection.Register(srv)

	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf(":%d", viper.GetInt("app.port")),
	)
	if err != nil {
		return fmt.Errorf("runner gRPC listen: %w", err)
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info().
			Int("port", viper.GetInt("app.port")).
			Bool("tls", viper.GetBool("app.tls.status")).
			Msg("Starting runner gRPC server")
		serverErrors <- srv.Serve(listener)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil {
			return fmt.Errorf("runner gRPC: %w", err)
		}
		return nil
	case sig := <-quit:
		log.Info().
			Str("signal", sig.String()).
			Msg("Received shutdown signal")

		signal.Stop(quit)
		cancel()

		shutdownTimeout := 30 * time.Second
		shutdownCtx, shutdownCancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer shutdownCancel()

		log.Info().
			Dur("timeout", shutdownTimeout).
			Msg("Gracefully shutting down runner gRPC server")

		stopped := make(chan struct{})
		go func() {
			srv.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			// ok
		case <-shutdownCtx.Done():
			srv.Stop()
		}

		select {
		case <-workerDone:
			// worker exited
		case <-shutdownCtx.Done():
			log.Warn().Msg("Worker did not exit before shutdown timeout")
		}

		var closeErr error
		if db != nil {
			closeErr = db.Close()
			if closeErr != nil {
				log.Error().
					Err(closeErr).
					Msg("Failed to close runner database")
			}
		}

		log.Info().Msg("Runner shutdown complete")
		if closeErr != nil {
			return fmt.Errorf("runner db close: %w", closeErr)
		}
	}

	return nil
}
