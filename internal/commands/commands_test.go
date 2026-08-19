package commands_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/useabstrax/abstrax/plugins/composer/internal/commands"
	"github.com/useabstrax/abstrax/plugins/composer/internal/install"
	"github.com/useabstrax/abstrax/plugins/composer/internal/layout"
	"github.com/useabstrax/abstrax/plugins/composer/internal/userx"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	out := filepath.Join(dir, "abstrax-composer")
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-o", out, "./cmd/abstrax-composer")
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, output)
	}
	return out
}

func TestPluginMetadata(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "plugin", "metadata")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() > 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
	var meta map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &meta); err != nil {
		t.Fatal(err)
	}
	if meta["name"] != "composer" {
		t.Fatalf("name = %v", meta["name"])
	}
	commandsList, _ := meta["commands"].([]any)
	names := map[string]bool{}
	for _, c := range commandsList {
		m := c.(map[string]any)
		names[m["name"].(string)] = true
	}
	for _, want := range []string{"setup", "self-update", "remove", "status", "configure", "run", "diagnose", "auth"} {
		if !names[want] {
			t.Fatalf("missing command %q in metadata", want)
		}
	}
}

func TestVersion(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "Plugin version:") {
		t.Fatalf("output = %q", stdout.String())
	}
}

func TestUnknownCommandExitCode(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "unknown-cmd")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error")
	}
	ee, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if ee.ExitCode() != 2 {
		t.Fatalf("exit code = %d, want 2", ee.ExitCode())
	}
}

func TestHelpShowsAbstraxCommand(t *testing.T) {
	root := commands.NewRootCmd()
	root.SetArgs([]string{"--help"})
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	got := stdout.String()
	if !strings.Contains(got, "abstrax composer [command]") {
		t.Fatalf("root help missing abstrax composer usage:\n%s", got)
	}
	if !strings.Contains(got, `"abstrax composer [command] --help"`) {
		t.Fatalf("root help missing abstrax composer help hint:\n%s", got)
	}
	if strings.Contains(got, "abstrax-composer") {
		t.Fatalf("root help still mentions binary name:\n%s", got)
	}
}

func TestSubcommandHelpShowsAbstraxCommand(t *testing.T) {
	root := commands.NewRootCmd()
	root.SetArgs([]string{"setup", "--help"})
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	got := stdout.String()
	if !strings.Contains(got, "abstrax composer setup") {
		t.Fatalf("setup help missing abstrax composer path:\n%s", got)
	}
	if strings.Contains(got, "abstrax-composer") {
		t.Fatalf("setup help still mentions binary name:\n%s", got)
	}
}

func TestJSONStreamMutuallyExclusive(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "version", "--json", "--json-stream")
	if err := cmd.Run(); err == nil {
		t.Fatal("expected error")
	}
}

func isolateLayout(t *testing.T) layout.Paths {
	t.Helper()
	dir := t.TempDir()
	original := layout.DefaultPaths
	ps := layout.Paths{
		ConfigPath:  filepath.Join(dir, "composer.json"),
		LibDir:      filepath.Join(dir, "lib"),
		PharName:    "composer.phar",
		WrapperPath: filepath.Join(dir, "bin", "composer"),
	}
	layout.DefaultPaths = ps
	t.Cleanup(func() { layout.DefaultPaths = original })
	return ps
}

func skipRoot(t *testing.T) {
	t.Helper()
	original := userx.RequireRoot
	userx.RequireRoot = func() error { return nil }
	t.Cleanup(func() { userx.RequireRoot = original })
}

func writePHP(t *testing.T, dir, name string) {
	t.Helper()
	script := `#!/bin/sh
if [ "$1" = "-r" ]; then echo "8.2.0"; exit 0; fi
if [ "$1" = "-m" ]; then printf "json\nmbstring\nxml\nzip\ncurl\n"; exit 0; fi
shift
echo "composer-args:$*"
`
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}

func composerServer(t *testing.T, phar []byte) *httptest.Server {
	t.Helper()
	sum := sha256.Sum256(phar)
	hash := hex.EncodeToString(sum[:])
	mux := http.NewServeMux()
	mux.HandleFunc("/versions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"stable":[{"path":"/download/2.8.0/composer.phar","version":"2.8.0"}]}`)
	})
	mux.HandleFunc("/download/2.8.0/composer.phar.sha256sum", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  composer.phar\n", hash)
	})
	mux.HandleFunc("/download/2.8.0/composer.phar", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(phar)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestSetupInstallsPharAndWrapper(t *testing.T) {
	skipRoot(t)
	ps := isolateLayout(t)
	binDir := t.TempDir()
	writePHP(t, binDir, "php")
	t.Setenv("PATH", binDir)
	t.Setenv("ABSTRAX_COMPOSER_PHP", "")

	phar := []byte("composer-phar-bytes")
	srv := composerServer(t, phar)
	original := commands.NewDownloader
	commands.NewDownloader = func() *install.Downloader {
		return &install.Downloader{BaseURL: srv.URL, HTTPClient: srv.Client()}
	}
	t.Cleanup(func() { commands.NewDownloader = original })

	root := commands.NewRootCmd()
	root.SetArgs([]string{"setup", "--yes"})
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(ps.PharPath())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(phar) {
		t.Fatalf("phar = %q", got)
	}
	if !install.IsManagedWrapper(ps.WrapperPath) {
		t.Fatal("wrapper not managed")
	}
}

func TestSetupDryRunDoesNotWrite(t *testing.T) {
	skipRoot(t)
	ps := isolateLayout(t)
	binDir := t.TempDir()
	writePHP(t, binDir, "php")
	t.Setenv("PATH", binDir)
	t.Setenv("ABSTRAX_COMPOSER_PHP", "")

	root := commands.NewRootCmd()
	root.SetArgs([]string{"setup", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ps.PharPath()); !os.IsNotExist(err) {
		t.Fatal("dry-run should not write phar")
	}
}

func TestConfigureSetsPHP(t *testing.T) {
	skipRoot(t)
	ps := isolateLayout(t)
	binDir := t.TempDir()
	writePHP(t, binDir, "php8.2")
	writePHP(t, binDir, "php")
	t.Setenv("PATH", binDir)
	t.Setenv("ABSTRAX_COMPOSER_PHP", "")

	root := commands.NewRootCmd()
	root.SetArgs([]string{"configure", "--php=php8.2"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(ps.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "php8.2") {
		t.Fatalf("config = %s", data)
	}
}

func TestRunForwardsComposerArgs(t *testing.T) {
	ps := isolateLayout(t)
	if err := os.MkdirAll(ps.LibDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ps.PharPath(), []byte("phar"), 0o644); err != nil {
		t.Fatal(err)
	}
	binDir := t.TempDir()
	writePHP(t, binDir, "php")
	t.Setenv("PATH", binDir)
	t.Setenv("ABSTRAX_COMPOSER_PHP", "")
	t.Setenv("SUDO_USER", "")

	root := commands.NewRootCmd()
	root.SetArgs([]string{"run", "install", "--no-dev", "--optimize-autoloader"})
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	// execx writes to os.Stdout, not cobra's writer
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRunRequiresComposerArgs(t *testing.T) {
	root := commands.NewRootCmd()
	root.SetArgs([]string{"run"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunDryRun(t *testing.T) {
	ps := isolateLayout(t)
	if err := os.MkdirAll(ps.LibDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ps.PharPath(), []byte("phar"), 0o644); err != nil {
		t.Fatal(err)
	}
	binDir := t.TempDir()
	writePHP(t, binDir, "php")
	t.Setenv("PATH", binDir)
	t.Setenv("ABSTRAX_COMPOSER_PHP", "")
	t.Setenv("SUDO_USER", "")

	root := commands.NewRootCmd()
	root.SetArgs([]string{"run", "--dry-run", "install", "--no-dev"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestStatusWithoutInstall(t *testing.T) {
	isolateLayout(t)
	binDir := t.TempDir()
	writePHP(t, binDir, "php")
	t.Setenv("PATH", binDir)
	t.Setenv("ABSTRAX_COMPOSER_PHP", "")

	root := commands.NewRootCmd()
	root.SetArgs([]string{"status"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRunMissingPhar(t *testing.T) {
	isolateLayout(t)
	binDir := t.TempDir()
	writePHP(t, binDir, "php")
	t.Setenv("PATH", binDir)
	t.Setenv("ABSTRAX_COMPOSER_PHP", "")

	root := commands.NewRootCmd()
	root.SetArgs([]string{"run", "install"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error")
	}
}
