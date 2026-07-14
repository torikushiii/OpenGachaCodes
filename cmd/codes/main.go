package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"opengachacodes/internal/buildinfo"
	"opengachacodes/internal/catalog"
	"opengachacodes/internal/config"
	"opengachacodes/internal/service"
)

func main() {
	configPath := flag.String("config", "config.yaml", "YAML configuration file")
	game := flag.String("game", "genshin", "game slug to collect")
	sourceName := flag.String("source", "all", "source adapter to collect, or all")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	ctx := context.Background()

	gameCatalog := catalog.New()
	configured := gameCatalog.Sources(
		&http.Client{Timeout: cfg.HTTP.Timeout},
		buildinfo.UserAgent(),
	)
	sources, err := gameCatalog.Select(*game, *sourceName, configured)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	result := service.CollectWithTimeout(ctx, sources, cfg.HTTP.Timeout)
	for _, sourceErr := range result.Errors {
		fmt.Fprintf(os.Stderr, "%s: %v\n", sourceErr.SourceID, sourceErr.Err)
	}
	if len(result.Codes) == 0 {
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result.Codes); err != nil {
		fmt.Fprintf(os.Stderr, "encode output: %v\n", err)
		os.Exit(1)
	}
}
