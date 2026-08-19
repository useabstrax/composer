package execx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// Request describes a Composer invocation.
type Request struct {
	PHP      string
	Phar     string
	Args     []string
	Dir      string
	User     string
	Stdout   io.Writer
	Stderr   io.Writer
	Stdin    io.Reader
	ExtraEnv []string
}

// CommandLine returns the argv that would be executed.
func (r Request) CommandLine() []string {
	argv := r.argv()
	return append([]string{argv[0]}, argv[1:]...)
}

func (r Request) argv() []string {
	php := r.PHP
	if php == "" {
		php = "php"
	}
	if r.User != "" && r.User != "root" && os.Geteuid() == 0 && !sameUser(r.User) {
		args := []string{"runuser", "-u", r.User, "--", php, r.Phar}
		return append(args, r.Args...)
	}
	args := []string{php, r.Phar}
	return append(args, r.Args...)
}

func sameUser(name string) bool {
	u, err := user.Current()
	if err != nil {
		return false
	}
	return u.Username == name
}

// Run executes Composer. Stdout/stderr default to the process streams.
func Run(ctx context.Context, req Request) error {
	argv := req.argv()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	if req.Dir != "" {
		cmd.Dir = req.Dir
	}
	cmd.Env = append(os.Environ(), req.ExtraEnv...)
	if req.Stdout != nil {
		cmd.Stdout = req.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if req.Stderr != nil {
		cmd.Stderr = req.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	if req.Stdin != nil {
		cmd.Stdin = req.Stdin
	} else {
		cmd.Stdin = os.Stdin
	}
	return cmd.Run()
}

// Capture runs Composer and returns combined-style stdout/stderr.
func Capture(ctx context.Context, req Request) (stdout, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer
	req.Stdout = &outBuf
	req.Stderr = &errBuf
	req.Stdin = nil
	err = Run(ctx, req)
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), err
}

// FormatCommand returns a display string for dry-run output.
func FormatCommand(req Request) string {
	parts := req.CommandLine()
	quoted := make([]string, len(parts))
	for i, p := range parts {
		if strings.ContainsAny(p, " \t\"'") {
			quoted[i] = fmt.Sprintf("%q", p)
		} else {
			quoted[i] = p
		}
	}
	return strings.Join(quoted, " ")
}
