package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/useabstrax/abstrax/plugins/composer/internal/install"
	"github.com/useabstrax/abstrax/plugins/composer/internal/output"
)

func newStatusCmd() *cobra.Command {
	var phpFlag, projectName string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Composer install state and the resolved PHP binary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := printer()
			ps := paths()
			action := "composer.status"

			phpRes, _, err := resolvePHP(cmd.Context(), phpFlag, projectName)
			if err != nil {
				p.Warn("%v", err)
			}

			phar := ps.PharPath()
			_, pharErr := os.Stat(phar)
			installed := pharErr == nil
			managed := install.IsManagedWrapper(ps.WrapperPath)
			composerVersion := ""
			if installed && phpRes.Binary != "" {
				composerVersion = probeComposerVersion(cmd.Context(), phpRes.Binary, phar)
			}

			cfg, cfgErr := loadPluginConfig()
			defaultPHP := "php"
			if cfgErr == nil {
				defaultPHP = cfg.PHPOrDefault()
			}

			data := map[string]any{
				"installed":        installed,
				"managed_wrapper":  managed,
				"phar_path":        phar,
				"wrapper_path":     ps.WrapperPath,
				"config_path":      ps.ConfigPath,
				"default_php":      defaultPHP,
				"php":              phpRes,
				"composer_version": composerVersion,
			}

			if !p.JSONMode && !p.JSONStream {
				p.Line("  Installed:         %s", yesNo(installed))
				p.Line("  Wrapper:           %s", ps.WrapperPath)
				p.Line("  Managed wrapper:   %s", yesNo(managed))
				p.Line("  Phar:              %s", phar)
				if composerVersion != "" {
					p.Line("  Composer version:  %s", composerVersion)
				}
				p.Line("  Default PHP:       %s", defaultPHP)
				if phpRes.Binary != "" {
					p.Line("  Resolved PHP:      %s (%s)", phpRes.Binary, phpRes.Source)
					if phpRes.Version != "" {
						p.Line("  PHP version:       %s", phpRes.Version)
					}
				}
			}

			summary := "Composer is not installed."
			if installed {
				summary = fmt.Sprintf("Composer is installed at %s.", ps.WrapperPath)
			}
			p.Print(output.Success(action, summary, data))
			return nil
		},
	}
	cmd.Flags().StringVar(&phpFlag, "php", "", "PHP binary to resolve for this status check")
	cmd.Flags().StringVar(&projectName, "project", "", "Resolve PHP from this Abstrax project")
	return cmd
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
