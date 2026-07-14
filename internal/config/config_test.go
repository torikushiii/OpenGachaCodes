package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte("database:\n  uri: mongodb://localhost:6000\n  name: opengachacodes\nhttp:\n  timeout: 12s\n")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.URI != "mongodb://localhost:6000" || cfg.HTTP.Timeout != 12*time.Second {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestLoadUsesDatabaseAndListenDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("database: {}\nhttp: {}\nscheduler: {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.URI != DefaultDatabaseURI || cfg.Database.Name != DefaultDatabaseName || cfg.HTTP.Listen != "127.0.0.1:8080" {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("database:\n  uri: mongodb://localhost:6000\n  name: test\ngame8:\n  pageURL: https://example.com\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unknown field error")
	}
}
