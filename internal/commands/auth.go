package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/useabstrax/abstrax/plugins/composer/internal/auth"
	"github.com/useabstrax/abstrax/plugins/composer/internal/output"
	"github.com/useabstrax/abstrax/plugins/composer/internal/userx"
)

func newAuthCmd() *cobra.Command {
	var (
		projectName  string
		userFlag     string
		githubToken  string
		httpHost     string
		httpUser     string
		httpPassword string
		removeKind   string
	)
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Show or update Composer authentication for a user",
		Long: `Show or update Composer auth.json for a user.

Tokens are written to ~/.config/composer/auth.json for the selected user
and are never printed in full. Use --show to confirm what is configured.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := printer()
			action := "composer.auth"

			username := strings.TrimSpace(userFlag)
			if username == "" && strings.TrimSpace(projectName) != "" {
				proj, err := loadProject(cmd.Context(), projectName)
				if err != nil {
					return err
				}
				if proj != nil && !userx.IsSharedOwner(proj.User) {
					username = proj.User
				}
			}
			if username == "" {
				asUser, err := userx.EffectiveRunUser("", true)
				if err != nil {
					return err
				}
				username = asUser
			}

			path, err := auth.PathForUser(username)
			if err != nil {
				return err
			}
			file, err := auth.Load(path)
			if err != nil {
				return err
			}

			changing := cmd.Flags().Changed("github-token") ||
				cmd.Flags().Changed("http-basic-host") ||
				cmd.Flags().Changed("remove")

			if !changing {
				redacted := file.Redacted()
				data := map[string]any{
					"user":             username,
					"path":             path,
					"github":           redacted.HasGithub(),
					"http_basic_hosts": redacted.HTTPBasicHosts(),
				}
				if !p.JSONMode && !p.JSONStream {
					p.Line("  User:     %s", username)
					p.Line("  Path:     %s", path)
					p.Line("  GitHub:   %s", yesNo(redacted.HasGithub()))
					if hosts := redacted.HTTPBasicHosts(); len(hosts) > 0 {
						p.Line("  HTTP basic hosts: %s", strings.Join(hosts, ", "))
					} else {
						p.Line("  HTTP basic hosts: none")
					}
				}
				p.Print(output.Success(action, fmt.Sprintf("Composer auth for %s.", username), data))
				return nil
			}

			if osEUIDRootRequired(username) {
				if err := userx.RequireRoot(); err != nil {
					return err
				}
			}

			if cmd.Flags().Changed("github-token") {
				if strings.TrimSpace(githubToken) == "" {
					file.ClearGithub()
				} else {
					file.SetGithubToken(githubToken)
				}
			}
			if cmd.Flags().Changed("http-basic-host") {
				host := strings.TrimSpace(httpHost)
				if host == "" {
					return fmt.Errorf("--http-basic-host is required when setting HTTP basic auth")
				}
				if strings.TrimSpace(httpUser) == "" || strings.TrimSpace(httpPassword) == "" {
					return fmt.Errorf("--username and --password are required with --http-basic-host")
				}
				file.SetHTTPBasic(host, httpUser, httpPassword)
			}
			if cmd.Flags().Changed("remove") {
				switch strings.ToLower(strings.TrimSpace(removeKind)) {
				case "github":
					file.ClearGithub()
				case "http-basic":
					file.ClearHTTPBasic(httpHost)
				default:
					return fmt.Errorf("--remove must be github or http-basic")
				}
			}

			if globals.DryRun {
				p.DryRun("write %s", path)
				p.Print(output.Success(action, "Would update Composer auth.", map[string]any{
					"user": username,
					"path": path,
				}))
				return nil
			}

			if err := auth.Save(path, file); err != nil {
				return err
			}
			p.Print(output.Success(action, fmt.Sprintf("Composer auth updated for %s.", username), map[string]any{
				"user":             username,
				"path":             path,
				"github":           file.HasGithub(),
				"http_basic_hosts": file.HTTPBasicHosts(),
			}))
			return nil
		},
	}
	cmd.Flags().StringVar(&projectName, "project", "", "Write auth.json for this Abstrax project user")
	cmd.Flags().StringVar(&userFlag, "user", "", "User whose Composer auth.json should be updated")
	cmd.Flags().StringVar(&githubToken, "github-token", "", "GitHub OAuth token (empty value removes it)")
	cmd.Flags().StringVar(&httpHost, "http-basic-host", "", "Host for HTTP basic auth (for example repo.packagist.com)")
	cmd.Flags().StringVar(&httpUser, "username", "", "HTTP basic username")
	cmd.Flags().StringVar(&httpPassword, "password", "", "HTTP basic password")
	cmd.Flags().StringVar(&removeKind, "remove", "", "Remove stored credentials: github or http-basic")
	return cmd
}

func osEUIDRootRequired(username string) bool {
	return username != userx.CurrentUsername()
}
