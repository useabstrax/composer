package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const Version = 1

// File is the plugin-owned /etc/abstrax/composer.json schema.
type File struct {
	Version int    `json:"version"`
	PHP     string `json:"php,omitempty"`
}

// Default returns an empty config that uses the unversioned php binary.
func Default() File {
	return File{Version: Version}
}

// Load reads config from path. Missing files yield Default.
func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return File{}, err
	}
	var cfg File
	if err := json.Unmarshal(data, &cfg); err != nil {
		return File{}, fmt.Errorf("invalid composer config: %w", err)
	}
	if cfg.Version == 0 {
		cfg.Version = Version
	}
	if cfg.Version != Version {
		return File{}, fmt.Errorf("unsupported composer.json version %d (want %d)", cfg.Version, Version)
	}
	cfg.PHP = strings.TrimSpace(cfg.PHP)
	return cfg, nil
}

// Save writes config atomically with mode 0640.
func Save(path string, cfg File) error {
	if cfg.Version == 0 {
		cfg.Version = Version
	}
	if cfg.Version != Version {
		return fmt.Errorf("unsupported composer.json version %d (want %d)", cfg.Version, Version)
	}
	cfg.PHP = strings.TrimSpace(cfg.PHP)
	if cfg.PHP == "php" {
		cfg.PHP = ""
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o640); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// PHPOrDefault returns the configured binary name, or "php".
func (c File) PHPOrDefault() string {
	if strings.TrimSpace(c.PHP) == "" {
		return "php"
	}
	return c.PHP
}
