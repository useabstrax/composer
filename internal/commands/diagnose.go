package commands

import (
	"context"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/useabstrax/abstrax/plugins/composer/internal/execx"
	"github.com/useabstrax/abstrax/plugins/composer/internal/output"
)

type diagnoseCheck struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Detail  string `json:"detail,omitempty"`
	Message string `json:"message,omitempty"`
}

func newDiagnoseCmd() *cobra.Command {
	var phpFlag, projectName string
	cmd := &cobra.Command{
		Use:   "diagnose",
		Short: "Check Composer, PHP, and common server prerequisites",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := printer()
			action := "composer.diagnose"
			ctx := cmd.Context()
			checks := []diagnoseCheck{}

			phpRes, proj, err := resolvePHP(ctx, phpFlag, projectName)
			if err != nil {
				checks = append(checks, diagnoseCheck{Name: "php", OK: false, Message: err.Error()})
			} else {
				detail := phpRes.Binary
				if phpRes.Version != "" {
					detail += " (" + phpRes.Version + ", " + string(phpRes.Source) + ")"
				}
				checks = append(checks, diagnoseCheck{Name: "php", OK: true, Detail: detail})
				for _, ext := range []string{"json", "mbstring", "xml", "zip", "curl"} {
					ok, msg := phpHasModule(ctx, phpRes.Binary, ext)
					checks = append(checks, diagnoseCheck{Name: "php-" + ext, OK: ok, Message: msg})
				}
			}

			phar, err := requirePhar()
			if err != nil {
				checks = append(checks, diagnoseCheck{Name: "composer", OK: false, Message: err.Error()})
			} else {
				ver := ""
				if phpRes.Binary != "" {
					ver = probeComposerVersion(ctx, phpRes.Binary, phar)
				}
				detail := phar
				if ver != "" {
					detail += " (" + ver + ")"
				}
				checks = append(checks, diagnoseCheck{Name: "composer", OK: true, Detail: detail})
			}

			for _, bin := range []string{"git", "unzip"} {
				path, lookErr := exec.LookPath(bin)
				if lookErr != nil {
					checks = append(checks, diagnoseCheck{
						Name:    bin,
						OK:      false,
						Message: bin + " not found on PATH (install it with abstrax package install " + bin + ")",
					})
					continue
				}
				checks = append(checks, diagnoseCheck{Name: bin, OK: true, Detail: path})
			}

			if proj != nil {
				checks = append(checks, diagnoseCheck{
					Name:   "project",
					OK:     true,
					Detail: proj.Name + " (" + proj.Path + ", user " + proj.User + ")",
				})
			}

			failed := 0
			for _, c := range checks {
				if !c.OK {
					failed++
				}
			}

			composerOut := ""
			if phar != "" && phpRes.Binary != "" && !globals.DryRun {
				stdout, stderr, runErr := execx.Capture(ctx, execx.Request{
					PHP:  phpRes.Binary,
					Phar: phar,
					Args: []string{"diagnose", "--no-ansi", "--no-interaction"},
				})
				composerOut = strings.TrimSpace(stdout + "\n" + stderr)
				if runErr != nil && composerOut == "" {
					composerOut = runErr.Error()
				}
			}

			if !p.JSONMode && !p.JSONStream {
				for _, c := range checks {
					mark := "ok"
					if !c.OK {
						mark = "FAIL"
					}
					line := c.Name + ": " + mark
					if c.Detail != "" {
						line += "  " + c.Detail
					}
					if c.Message != "" {
						line += "  " + c.Message
					}
					p.Line("%s", line)
				}
				if composerOut != "" {
					p.Line("")
					p.Line("composer diagnose:")
					p.Line("%s", composerOut)
				}
			}

			summary := "Composer environment looks OK."
			if failed > 0 {
				summary = "Composer environment has issues."
			}
			p.Print(output.Success(action, summary, map[string]any{
				"checks":            checks,
				"failed":            failed,
				"php":               phpRes,
				"composer_diagnose": composerOut,
			}))
			return nil
		},
	}
	cmd.Flags().StringVar(&phpFlag, "php", "", "PHP binary to diagnose")
	cmd.Flags().StringVar(&projectName, "project", "", "Include checks for this Abstrax project")
	return cmd
}

func phpHasModule(ctx context.Context, binary, name string) (bool, string) {
	cmd := exec.CommandContext(ctx, binary, "-m")
	out, err := cmd.Output()
	if err != nil {
		return false, "could not list PHP modules"
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.EqualFold(strings.TrimSpace(line), name) {
			return true, ""
		}
	}
	return false, "PHP extension " + name + " is not loaded"
}

func probeComposerVersion(ctx context.Context, phpBin, phar string) string {
	stdout, _, err := execx.Capture(ctx, execx.Request{
		PHP:  phpBin,
		Phar: phar,
		Args: []string{"--version", "--no-ansi", "--no-interaction"},
	})
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(strings.Split(stdout, "\n")[0])
	return line
}
