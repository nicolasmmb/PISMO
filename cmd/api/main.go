package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nicolasmmb/pismo-challenge/internal/adapter/clock"
	loggeradapter "github.com/nicolasmmb/pismo-challenge/internal/adapter/logger"
	"github.com/nicolasmmb/pismo-challenge/internal/config"
)

func main() {
	cfg := config.Load()
	log := loggeradapter.New()
	clk := clock.SystemClock{}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("starting service", map[string]any{"port": cfg.Port, "db_driver": cfg.DBDriver})

	if err := run(ctx, cfg, log, clk); err != nil {
		slog.Error("service failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config, log loggeradapter.SlogLogger, clk clock.SystemClock) error {
	// TODO: wire repositories, use cases, HTTP server
	<-ctx.Done()
	log.Info("shutdown complete", nil)
	return nil
}
