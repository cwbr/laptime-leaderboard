package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ingesthttp "github.com/chris/laptime-leaderboard/internal/adapter/ingest/http"
	"github.com/chris/laptime-leaderboard/internal/adapter/repository/sqlite"
	"github.com/chris/laptime-leaderboard/internal/adapter/web"
	"github.com/chris/laptime-leaderboard/internal/config"
	"github.com/chris/laptime-leaderboard/internal/domain"
)

func main() {
	configPath := flag.String("config", "", "path to config file (optional)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Database
	db, err := sqlite.New(cfg.DatabasePath, logger)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Repositories
	lapRepo := sqlite.NewLapRepo(db)
	serverRepo := sqlite.NewServerRepo(db)
	gameRepo := sqlite.NewGameRepo(db)
	trackRepo := sqlite.NewTrackRepo(db)
	carRepo := sqlite.NewCarRepo(db)
	playerRepo := sqlite.NewPlayerRepo(db)

	// Core service
	svc := domain.NewService(lapRepo, serverRepo, gameRepo, trackRepo, carRepo, playerRepo, logger)

	// HTTP mux — single mux serves both the ingest API and the web UI
	mux := http.NewServeMux()

	// Ingest adapter (receives POSTs from game server plugins)
	ingestHandler := ingesthttp.NewHandler(svc, logger)
	ingestHandler.RegisterRoutes(mux)

	// Web UI adapter (serves HTML leaderboard)
	webHandler := web.NewHandler(svc, logger)
	webHandler.RegisterRoutes(mux)

	// HTTP server
	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	logger.Info("starting server", "addr", cfg.ListenAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
