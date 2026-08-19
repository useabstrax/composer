package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/useabstrax/abstrax/plugins/composer/internal/install"
	"github.com/useabstrax/abstrax/plugins/composer/internal/output"
	"github.com/useabstrax/abstrax/plugins/composer/internal/userx"
)

func newRemoveCmd() *cobra.Command {
	var purge bool
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove the Composer binary installed by this plugin",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := userx.RequireRoot(); err != nil {
				return err
			}
			p := printer()
			ps := paths()
			action := "composer.remove"

			if install.WrapperExists(ps.WrapperPath) && !install.IsManagedWrapper(ps.WrapperPath) {
				return fmt.Errorf("%s exists but was not installed by Abstrax; remove it manually", ps.WrapperPath)
			}

			if !globals.DryRun && !globals.Yes {
				ok, err := confirm("Remove the Composer binary installed by Abstrax?")
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("cancelled")
				}
			}

			if globals.DryRun {
				p.DryRun("remove %s", ps.WrapperPath)
				p.DryRun("remove %s", ps.PharPath())
				if purge {
					p.DryRun("remove %s", ps.ConfigPath)
				}
				p.Print(output.Success(action, "Would remove Composer.", map[string]any{
					"wrapper_path": ps.WrapperPath,
					"phar_path":    ps.PharPath(),
					"purge":        purge,
				}))
				return nil
			}

			removed := []string{}
			if err := os.Remove(ps.WrapperPath); err != nil && !os.IsNotExist(err) {
				return err
			} else if err == nil {
				removed = append(removed, ps.WrapperPath)
			}
			if err := os.Remove(ps.PharPath()); err != nil && !os.IsNotExist(err) {
				return err
			} else if err == nil {
				removed = append(removed, ps.PharPath())
			}
			_ = os.Remove(ps.LibDir)
			if purge {
				if err := os.Remove(ps.ConfigPath); err != nil && !os.IsNotExist(err) {
					return err
				} else if err == nil {
					removed = append(removed, ps.ConfigPath)
				}
			}

			p.Print(output.Success(action, "Composer removed.", map[string]any{
				"removed": removed,
				"purge":   purge,
			}))
			return nil
		},
	}
	cmd.Flags().BoolVar(&purge, "purge", false, "Also remove /etc/abstrax/composer.json")
	return cmd
}
