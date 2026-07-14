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

	"opengachacodes/internal/api"
	"opengachacodes/internal/buildinfo"
	"opengachacodes/internal/catalog"
	"opengachacodes/internal/config"
	mongodbstore "opengachacodes/internal/repository/mongodb"
	"opengachacodes/internal/scheduler"
	"opengachacodes/internal/service"
)

func main() {
	configPath := flag.String("config", "config.yaml", "YAML configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	setupCtx, cancelSetup := context.WithTimeout(rootCtx, cfg.HTTP.Timeout)
	store, err := mongodbstore.Connect(setupCtx, cfg.Database.URI, cfg.Database.Name)
	cancelSetup()
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := store.Close(ctx); err != nil {
			logger.Error("database disconnect failed", "error", err)
		}
	}()

	gameCatalog := catalog.New()
	setupCtx, cancelSetup = context.WithTimeout(rootCtx, cfg.HTTP.Timeout)
	if err := store.EnsureIndexes(setupCtx); err == nil {
		err = store.RenameGameSlug(setupCtx, "genshin-impact", "genshin")
	}
	if err == nil {
		err = store.RenameGameSlug(setupCtx, "honkai-star-rail", "starrail")
	}
	if err == nil {
		err = store.EnsureGames(setupCtx, gameCatalog.Games())
	}
	if err == nil {
		err = store.DeleteCodes(setupCtx, "genshin", []string{"YUANSHEN"})
	}
	cancelSetup()
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: cfg.HTTP.Timeout}
	sources := gameCatalog.Sources(client, buildinfo.UserAgent())
	runner := service.Runner{Store: store, Sources: sources, SourceTimeout: cfg.HTTP.Timeout}
	runCollection := func(parent context.Context) {
		result := runner.Run(parent)
		for _, sourceErr := range result.Errors {
			logger.Error("collection issue", "source", sourceErr.SourceID, "error", sourceErr.Err)
		}
		logger.Info("collection finished", "codes", len(result.Codes), "issues", len(result.Errors))
	}

	go (scheduler.Scheduler{
		Location: cfg.Scheduler.Location,
		Run:      runCollection,
		Logger:   logger,
	}).Start(rootCtx)

	server := &http.Server{
		Addr:              cfg.HTTP.Listen,
		Handler:           api.Handler{Store: store},
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.HTTP.Timeout,
		WriteTimeout:      cfg.HTTP.Timeout,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("API server listening", "address", cfg.HTTP.Listen)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-rootCtx.Done():
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("API server failed", "error", err)
		}
		stop()
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("API shutdown failed", "error", err)
	}
}
