package php

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/useabstrax/abstrax/plugins/composer/sdk/abstrax"
)

const (
	// EnvPHP overrides the configured default PHP binary for one invocation.
	EnvPHP = "ABSTRAX_COMPOSER_PHP"
)

// Source explains how the PHP binary was chosen.
type Source string

const (
	SourceFlag    Source = "flag"
	SourceEnv     Source = "env"
	SourceProject Source = "project"
	SourceConfig  Source = "config"
	SourceDefault Source = "default"
)

// Resolution is the chosen PHP CLI binary.
type Resolution struct {
	Requested string `json:"requested"`
	Binary    string `json:"binary"`
	Source    Source `json:"source"`
	Version   string `json:"version,omitempty"`
}

// Options control PHP resolution.
type Options struct {
	Flag      string
	ConfigPHP string
	Project   *abstrax.Project
}

var lookPath = exec.LookPath

// Resolve chooses a PHP CLI binary.
//
// Order: --php flag, ABSTRAX_COMPOSER_PHP, project runtime, config, "php".
func Resolve(opts Options) (Resolution, error) {
	if bin := strings.TrimSpace(opts.Flag); bin != "" {
		return resolveNamed(bin, SourceFlag)
	}
	if bin := strings.TrimSpace(os.Getenv(EnvPHP)); bin != "" {
		return resolveNamed(bin, SourceEnv)
	}
	if opts.Project != nil && strings.EqualFold(strings.TrimSpace(opts.Project.Runtime.Type), "php") {
		bin := ResolveCLI(opts.Project.Runtime.Version)
		res, err := resolveNamed(bin, SourceProject)
		if err == nil {
			return res, nil
		}
		return Resolution{}, fmt.Errorf("project PHP %s: %w", opts.Project.Runtime.Version, err)
	}
	if bin := strings.TrimSpace(opts.ConfigPHP); bin != "" && bin != "php" {
		return resolveNamed(bin, SourceConfig)
	}
	return resolveNamed("php", SourceDefault)
}

func resolveNamed(name string, source Source) (Resolution, error) {
	path, err := locate(name)
	if err != nil {
		return Resolution{}, err
	}
	return Resolution{Requested: name, Binary: path, Source: source}, nil
}

func locate(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("PHP binary is empty")
	}
	if strings.Contains(name, "/") {
		st, err := os.Stat(name)
		if err != nil {
			return "", fmt.Errorf("PHP binary %q not found", name)
		}
		if st.IsDir() {
			return "", fmt.Errorf("PHP binary %q is a directory", name)
		}
		return name, nil
	}
	path, err := lookPath(name)
	if err != nil {
		return "", fmt.Errorf("PHP binary %q not found on PATH", name)
	}
	return path, nil
}

// ResolveCLI returns the best PHP CLI binary for a project runtime version.
func ResolveCLI(version string) string {
	version = strings.TrimSpace(version)
	candidates := []string{}
	if version != "" {
		compact := strings.ReplaceAll(version, ".", "")
		candidates = append(candidates,
			"php"+version,
			filepath.Join("/opt/remi", "php"+compact, "root/usr/bin/php"),
			filepath.Join("/usr/bin", "php"+version),
		)
	}
	candidates = append(candidates, "php")

	for _, c := range candidates {
		if path, err := lookPath(c); err == nil {
			return path
		}
		if strings.Contains(c, "/") {
			if st, err := os.Stat(c); err == nil && !st.IsDir() {
				return c
			}
		}
	}
	return "php"
}

// ProbeVersion runs php -r 'echo PHP_VERSION;' when possible.
func ProbeVersion(ctx context.Context, binary string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, "-r", "echo PHP_VERSION;")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("running %s: %s", binary, msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Validate checks that name exists and looks like PHP.
func Validate(ctx context.Context, name string) (Resolution, error) {
	res, err := resolveNamed(name, SourceFlag)
	if err != nil {
		return Resolution{}, err
	}
	ver, err := ProbeVersion(ctx, res.Binary)
	if err != nil {
		return Resolution{}, fmt.Errorf("%s does not look like a PHP CLI: %w", name, err)
	}
	res.Version = ver
	return res, nil
}
