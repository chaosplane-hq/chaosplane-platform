package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/server"
)

func main() {
	ctx := context.Background()

	srv, err := server.InitializeServer(ctx)
	if err != nil {
		slog.Error("failed to initialize server", "error", err)
		os.Exit(1)
	}
	defer srv.Pool.Close()

	httpServer := &http.Server{
		Addr:              ":" + srv.Config.Port,
		Handler:           srv.Router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", srv.Config.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}
}
