package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/useabstrax/abstrax/plugins/composer/internal/install"
	"github.com/useabstrax/abstrax/plugins/composer/internal/output"
	"github.com/useabstrax/abstrax/plugins/composer/internal/php"
	"github.com/useabstrax/abstrax/plugins/composer/internal/userx"
)

// NewDownloader creates the Composer downloader. Tests replace this.
var NewDownloader = func() *install.Downloader {
	return &install.Downloader{}
}

func newSetupCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Download Composer, verify it, and install it globally",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := userx.RequireRoot(); err != nil {
				return err
			}
			return runInstall(cmd.Context(), "composer.setup", force, false)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Replace an existing Composer binary even if it was not installed by this plugin")
	return cmd
}

func newSelfUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "self-update",
		Short: "Update the installed Composer phar to the latest stable version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := userx.RequireRoot(); err != nil {
				return err
			}
			if _, err := requirePhar(); err != nil {
				return err
			}
			return runInstall(cmd.Context(), "composer.self_update", true, true)
		},
	}
}

func runInstall(ctx context.Context, action string, force, selfUpdate bool) error {
	p := printer()
	ps := paths()

	if install.WrapperExists(ps.WrapperPath) && !install.IsManagedWrapper(ps.WrapperPath) && !force {
		return fmt.Errorf("%s already exists and was not installed by Abstrax; re-run with --force to replace it", ps.WrapperPath)
	}

	cfg, err := loadPluginConfig()
	if err != nil {
		return err
	}
	phpRes, err := php.Resolve(php.Options{ConfigPHP: cfg.PHP})
	if err != nil {
		return fmt.Errorf("resolving PHP: %w", err)
	}
	if ver, probeErr := php.ProbeVersion(ctx, phpRes.Binary); probeErr != nil {
		return fmt.Errorf("resolving PHP: %w", probeErr)
	} else {
		phpRes.Version = ver
	}

	if globals.DryRun {
		p.DryRun("download latest stable Composer from getcomposer.org")
		p.DryRun("write %s", ps.PharPath())
		p.DryRun("write wrapper %s using %s", ps.WrapperPath, phpRes.Binary)
		p.Print(output.Success(action, "Would install Composer.", map[string]any{
			"phar_path":    ps.PharPath(),
			"wrapper_path": ps.WrapperPath,
			"php":          phpRes,
		}))
		return nil
	}

	p.Progress(action, "download", "Downloading Composer")
	dl := NewDownloader()
	downloadCtx := ctx
	if downloadCtx == nil {
		downloadCtx = context.Background()
	}
	downloadCtx, cancel := context.WithTimeout(downloadCtx, 90*time.Second)
	defer cancel()

	release, err := dl.LatestStable(downloadCtx)
	if err != nil {
		return err
	}
	p.Verbose("Composer %s sha256 %s", release.Version, release.SHA256)

	p.Progress(action, "install", "Installing Composer")
	if err := install.WritePhar(ps, release.Phar); err != nil {
		return fmt.Errorf("writing Composer phar: %w", err)
	}
	if err := os.Chmod(ps.PharPath(), 0o644); err != nil {
		return err
	}
	if err := install.WriteWrapper(ps, phpRes.Binary); err != nil {
		return fmt.Errorf("writing Composer wrapper: %w", err)
	}

	summary := fmt.Sprintf("Composer %s installed at %s.", release.Version, ps.WrapperPath)
	if selfUpdate {
		summary = fmt.Sprintf("Composer updated to %s.", release.Version)
	}
	p.Print(output.Success(action, summary, map[string]any{
		"version":      release.Version,
		"phar_path":    ps.PharPath(),
		"wrapper_path": ps.WrapperPath,
		"sha256":       release.SHA256,
		"php":          phpRes,
	}))
	return nil
}
