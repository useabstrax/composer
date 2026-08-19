package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/useabstrax/abstrax/plugins/composer/internal/config"
)

func TestLoadMissingReturnsDefault(t *testing.T) {
	cfg, err := config.Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PHPOrDefault() != "php" {
		t.Fatalf("php = %q", cfg.PHPOrDefault())
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "composer.json")
	cfg := config.File{Version: 1, PHP: "php8.2"}
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	got, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.PHP != "php8.2" {
		t.Fatalf("php = %q", got.PHP)
	}
}

func TestSaveNormalizesDefaultPHP(t *testing.T) {
	path := filepath.Join(t.TempDir(), "composer.json")
	if err := config.Save(path, config.File{PHP: "php"}); err != nil {
		t.Fatal(err)
	}
	got, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.PHP != "" {
		t.Fatalf("php = %q, want empty", got.PHP)
	}
}

func TestLoadRejectsUnknownVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "composer.json")
	if err := os.WriteFile(path, []byte("{\n  \"version\": 99\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error")
	}
}
