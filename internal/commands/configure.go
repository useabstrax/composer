package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/useabstrax/abstrax/plugins/composer/internal/config"
	"github.com/useabstrax/abstrax/plugins/composer/internal/install"
	"github.com/useabstrax/abstrax/plugins/composer/internal/output"
	"github.com/useabstrax/abstrax/plugins/composer/internal/php"
	"github.com/useabstrax/abstrax/plugins/composer/internal/userx"
)

func newConfigureCmd() *cobra.Command {
	var phpFlag string
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Show or set the default PHP binary for Composer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := printer()
			ps := paths()
			action := "composer.configure"
			cfg, err := loadPluginConfig()
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("php") {
				data := map[string]any{
					"config_path": ps.ConfigPath,
					"php":         cfg.PHPOrDefault(),
				}
				if !p.JSONMode && !p.JSONStream {
					p.Line("  Config:  %s", ps.ConfigPath)
					p.Line("  PHP:     %s", cfg.PHPOrDefault())
				}
				p.Print(output.Success(action, fmt.Sprintf("Default PHP is %s.", cfg.PHPOrDefault()), data))
				return nil
			}

			if err := userx.RequireRoot(); err != nil {
				return err
			}

			requested := phpFlag
			if requested == "" {
				requested = "php"
			}
			res, err := php.Validate(cmd.Context(), requested)
			if err != nil {
				return err
			}

			cfg.PHP = requested
			if requested == "php" {
				cfg.PHP = ""
			}

			if globals.DryRun {
				p.DryRun("set default PHP to %s (%s)", res.Requested, res.Binary)
				if install.IsManagedWrapper(ps.WrapperPath) {
					p.DryRun("rewrite wrapper %s", ps.WrapperPath)
				}
				p.Print(output.Success(action, "Would update Composer PHP default.", map[string]any{
					"php": res,
				}))
				return nil
			}

			if err := config.Save(ps.ConfigPath, cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			if install.IsManagedWrapper(ps.WrapperPath) {
				if err := install.WriteWrapper(ps, res.Binary); err != nil {
					return fmt.Errorf("rewriting Composer wrapper: %w", err)
				}
			}

			summary := fmt.Sprintf("Default PHP set to %s.", res.Requested)
			p.Print(output.Success(action, summary, map[string]any{
				"config_path": ps.ConfigPath,
				"php":         res,
			}))
			return nil
		},
	}
	cmd.Flags().StringVar(&phpFlag, "php", "", "PHP binary to use when no project is given (for example php8.2). Pass php to reset.")
	return cmd
}
