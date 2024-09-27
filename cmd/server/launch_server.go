package main

import (
	"context"
	"photos/internal/handlers"

	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"photos/internal/routes"
	"syscall"
	"time"
)

func main() {
	cfg, err := handlers.New()
	if err != nil {
		log.Fatalf("error initialising handlers config: %v\n", err)
	}

	server := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%d", cfg.Server.Port),
		Handler:        routes.Service(cfg),
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
				log.Fatalf("graceful shutdown timed out: %v\n", shutdownCtx.Err())
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatalf("error shutting down server: %v\n", err)
		}
		shutdownCtxCancel()
		serverCtxCancel()
	}()

	log.Printf("starting server on %s\n", server.Addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("error starting server: %v\n", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
