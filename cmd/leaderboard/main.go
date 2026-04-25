package main

import (
	"context"
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
	"github.com/chris/laptime-leaderboard/internal/admin"
	"github.com/chris/laptime-leaderboard/internal/config"
	"github.com/chris/laptime-leaderboard/internal/domain"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load(os.Getenv("LEADERBOARD_CONFIG"))
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
	serverRepo := sqlite.NewServerRepo(db)
	gameRepo := sqlite.NewGameRepo(db)

	// Route subcommand
	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "serve":
		runServe(cfg, db, gameRepo, serverRepo, logger)
	case "admin":
		admin.Run(os.Args[2:], serverRepo, gameRepo)
	default:
		fmt.Fprintf(os.Stderr, "Usage: leaderboard <serve|admin> [args]\n")
		os.Exit(1)
	}
}

func runServe(cfg config.Config, db *sqlite.DB, gameRepo *sqlite.GameRepo, serverRepo *sqlite.ServerRepo, logger *slog.Logger) {
	lapRepo := sqlite.NewLapRepo(db)
	trackRepo := sqlite.NewTrackRepo(db)
	carRepo := sqlite.NewCarRepo(db)
	playerRepo := sqlite.NewPlayerRepo(db)

	svc := domain.NewService(lapRepo, serverRepo, gameRepo, trackRepo, carRepo, playerRepo, logger)

	mux := http.NewServeMux()

	ingestHandler := ingesthttp.NewHandler(svc, logger)
	ingestHandler.RegisterRoutes(mux)

	webHandler := web.NewHandler(svc, logger)
	webHandler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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
