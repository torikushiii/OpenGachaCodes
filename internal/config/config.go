package config

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Database  Database  `yaml:"database"`
	HTTP      HTTP      `yaml:"http"`
	Scheduler Scheduler `yaml:"scheduler"`
}

type Database struct {
	URI  string `yaml:"uri"`
	Name string `yaml:"name"`
}

type HTTP struct {
	Timeout    time.Duration `yaml:"-"`
	RawTimeout string        `yaml:"timeout"`
	Listen     string        `yaml:"listen"`
}

type Scheduler struct {
	LocationName string         `yaml:"location"`
	Location     *time.Location `yaml:"-"`
}

const (
	DefaultDatabaseURI  = "mongodb://localhost:27017"
	DefaultDatabaseName = "opengachacodes"
)

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if cfg.Database.URI == "" {
		cfg.Database.URI = DefaultDatabaseURI
	}
	if cfg.Database.Name == "" {
		cfg.Database.Name = DefaultDatabaseName
	}
	if cfg.HTTP.RawTimeout == "" {
		cfg.HTTP.RawTimeout = "15s"
	}
	cfg.HTTP.Timeout, err = time.ParseDuration(cfg.HTTP.RawTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("parse http timeout: %w", err)
	}
	if cfg.HTTP.Timeout <= 0 {
		return Config{}, fmt.Errorf("http timeout must be positive")
	}
	if cfg.HTTP.Listen == "" {
		cfg.HTTP.Listen = "127.0.0.1:8080"
	}
	if cfg.Scheduler.LocationName == "" || cfg.Scheduler.LocationName == "Local" {
		cfg.Scheduler.LocationName = "Local"
		cfg.Scheduler.Location = time.Local
	} else {
		cfg.Scheduler.Location, err = time.LoadLocation(cfg.Scheduler.LocationName)
		if err != nil {
			return Config{}, fmt.Errorf("load scheduler location: %w", err)
		}
	}
	return cfg, nil
}
