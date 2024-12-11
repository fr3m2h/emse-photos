package main

import (
	"context"

	"fmt"
	"net/http"
	"os"
	"os/signal"
	"photos/pkg/config"
	"photos/pkg/handlers"
	"photos/pkg/routes"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Millisecond
	cfg := config.Load()

	dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), 8*time.Second)
	currentTime := time.Now()
	if cfg.DevMode.Enabled {
		cfg.Logger.Info().Msg("using dev database")
	} else {
		cfg.Logger.Info().Msg("using prod database")
	}
	err := cfg.DB.PingContext(dbCtx)
	if err != nil {
		cfg.Logger.Fatal().Err(err).Msg("failed to ping database")
	}
	dbCtxCancel()
	cfg.Logger.Info().Dur("latency", time.Since(currentTime)).Msg("pinged database")

	server := &http.Server{
		Addr:           fmt.Sprintf("127.0.0.1:%d", cfg.Server.Port),
		Handler:        routes.Service(handlers.Config(cfg)),
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	serverCtx, serverCtxCancel := context.WithCancel(context.Background())
	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, shutdownCtxCancel := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				cfg.Logger.Fatal().Err(err).Msg("graceful shutdown timed out")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			cfg.Logger.Fatal().Err(err).Msg("failed to shut down server")
		}
		err = cfg.DB.Close()
		if err != nil {
			cfg.Logger.Fatal().Err(err).Msg("failed to close database connection")

		}
		shutdownCtxCancel()
		serverCtxCancel()
	}()

	cfg.Logger.Info().Str("address", server.Addr).Msg("starting server")
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		cfg.Logger.Fatal().Err(err).Msg("failed to start server")
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
