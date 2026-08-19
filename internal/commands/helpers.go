package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/useabstrax/abstrax/plugins/composer/internal/config"
	"github.com/useabstrax/abstrax/plugins/composer/internal/layout"
	"github.com/useabstrax/abstrax/plugins/composer/internal/php"
	"github.com/useabstrax/abstrax/plugins/composer/sdk/abstrax"
)

func paths() layout.Paths {
	return layout.DefaultPaths
}

func loadPluginConfig() (config.File, error) {
	return config.Load(paths().ConfigPath)
}

func confirm(prompt string) (bool, error) {
	if globals.Yes {
		return true, nil
	}
	fmt.Fprintf(os.Stderr, "%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes", nil
}

func loadProject(ctx context.Context, name string) (*abstrax.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil
	}
	client, err := abstrax.New()
	if err != nil {
		return nil, err
	}
	resp, err := client.Project(ctx, name)
	if err != nil {
		return nil, err
	}
	proj := resp.Project
	return &proj, nil
}

func resolvePHP(ctx context.Context, flagPHP, projectName string) (php.Resolution, *abstrax.Project, error) {
	cfg, err := loadPluginConfig()
	if err != nil {
		return php.Resolution{}, nil, err
	}
	proj, err := loadProject(ctx, projectName)
	if err != nil {
		return php.Resolution{}, nil, err
	}
	res, err := php.Resolve(php.Options{
		Flag:      flagPHP,
		ConfigPHP: cfg.PHP,
		Project:   proj,
	})
	if err != nil {
		return php.Resolution{}, proj, err
	}
	if ver, probeErr := php.ProbeVersion(ctx, res.Binary); probeErr == nil {
		res.Version = ver
	}
	return res, proj, nil
}

func requirePhar() (string, error) {
	path := paths().PharPath()
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return "", fmt.Errorf("Composer is not installed; run sudo abstrax composer setup")
	}
	return path, nil
}

func workingDir(flagPath string, proj *abstrax.Project) (string, error) {
	if strings.TrimSpace(flagPath) != "" {
		return flagPath, nil
	}
	if proj != nil && strings.TrimSpace(proj.Path) != "" {
		return proj.Path, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolving working directory: %w", err)
	}
	return cwd, nil
}
