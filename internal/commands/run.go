package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/useabstrax/abstrax/plugins/composer/internal/execx"
	"github.com/useabstrax/abstrax/plugins/composer/internal/output"
	"github.com/useabstrax/abstrax/plugins/composer/internal/userx"
)

func newRunCmd() *cobra.Command {
	var (
		projectName string
		pathFlag    string
		phpFlag     string
		userFlag    string
		allowRoot   bool
	)
	cmd := &cobra.Command{
		Use:   "run [composer-args...]",
		Short: "Run Composer with the resolved PHP binary",
		Long: `Run Composer with the resolved PHP binary.

Put Abstrax flags before the Composer command:

  abstrax composer run install
  abstrax composer run install --no-dev --optimize-autoloader
  abstrax composer run --project=myapp install --no-dev
  abstrax composer run --php=php8.2 update

Abstrax globals such as --verbose, --quiet, and --dry-run are parsed as
Abstrax flags. To pass those flags through to Composer, put them after --:

  abstrax composer run -- install --dry-run`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := printer()
			action := "composer.run"

			phpRes, proj, err := resolvePHP(cmd.Context(), phpFlag, projectName)
			if err != nil {
				return err
			}
			phar, err := requirePhar()
			if err != nil {
				return err
			}
			dir, err := workingDir(pathFlag, proj)
			if err != nil {
				return err
			}

			runUser := userFlag
			if runUser == "" && proj != nil && !userx.IsSharedOwner(proj.User) {
				runUser = proj.User
			}
			asUser, err := userx.EffectiveRunUser(runUser, allowRoot)
			if err != nil {
				return err
			}

			req := execx.Request{
				PHP:  phpRes.Binary,
				Phar: phar,
				Args: args,
				Dir:  dir,
				User: asUser,
			}

			if globals.DryRun {
				p.DryRun("cd %s", dir)
				p.DryRun("%s", execx.FormatCommand(req))
				p.Print(output.Success(action, "Would run Composer.", map[string]any{
					"php":           phpRes,
					"working_dir":   dir,
					"user":          asUser,
					"composer_args": args,
					"command":       execx.FormatCommand(req),
				}))
				return nil
			}

			p.Verbose("PHP %s (%s)", phpRes.Binary, phpRes.Source)
			p.Verbose("user %s cwd %s", asUser, dir)

			if p.JSONMode || p.JSONStream {
				stdout, stderr, runErr := execx.Capture(cmd.Context(), req)
				if runErr != nil {
					msg := strings.TrimSpace(stderr)
					if msg == "" {
						msg = runErr.Error()
					}
					return fmt.Errorf("composer: %s", msg)
				}
				p.Print(output.Success(action, "Composer finished.", map[string]any{
					"php":           phpRes,
					"working_dir":   dir,
					"user":          asUser,
					"composer_args": args,
					"stdout":        stdout,
					"stderr":        stderr,
				}))
				return nil
			}

			if err := execx.Run(cmd.Context(), req); err != nil {
				return fmt.Errorf("composer: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().SetInterspersed(false)
	cmd.Flags().StringVar(&projectName, "project", "", "Abstrax project to use for path, user, and PHP version")
	cmd.Flags().StringVar(&pathFlag, "path", "", "Working directory (default: current directory, or the project path)")
	cmd.Flags().StringVar(&phpFlag, "php", "", "PHP binary for this invocation")
	cmd.Flags().StringVar(&userFlag, "user", "", "User to run Composer as")
	cmd.Flags().BoolVar(&allowRoot, "allow-root", false, "Allow running Composer as root")
	return cmd
}
