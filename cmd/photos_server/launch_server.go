package main

import (
	"context"
	"log/slog"

	"fmt"
	"net/http"
	"os"
	"os/signal"
	"photos/internal/config"
	"photos/internal/handlers"
	"photos/internal/routes"
	"syscall"
	"time"

	slogmulti "github.com/samber/slog-multi"
)

func main() {
	logFile, err := os.Create("logs")
	if err != nil {
		slog.Error("error creating file for logs", "error", err)
	}
	defer func() {
		_ = logFile.Close()
	}()
	logger := slog.New(
		slogmulti.Fanout(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}),
			slog.NewJSONHandler(logFile, &slog.HandlerOptions{}),
		),
	)

	cfg, err := config.Load(logger)
	if err != nil {
		logger.Error("error initialising handlers config", "error", err)
		os.Exit(1)
	}

	dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), 8*time.Second)
	t := time.Now()
	err = cfg.DB.PingContext(dbCtx)
	if err != nil {
		logger.Error("error pinging to database", "error", err)
		os.Exit(1)
	}
	dbCtxCancel()
	if cfg.DevMode.Enabled {
		logger.Info("using dev database")
	} else {
		logger.Info("using prod database")
	}
	logger.Info("pinged database", "latency", time.Since(t).String())

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
				logger.Error("graceful shutdown timed out", "error", shutdownCtx.Err())
				os.Exit(1)
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error("error shutting down server", "error", err)
			os.Exit(1)
		}
		err = cfg.DB.Close()
		if err != nil {
			logger.Error("error closing database", "error", err)
			os.Exit(1)

		}
		shutdownCtxCancel()
		serverCtxCancel()
	}()

	logger.Info("starting server", "address", server.Addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Error("error starting server", "error", err)
		os.Exit(1)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
