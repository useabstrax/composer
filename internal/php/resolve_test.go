package php_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/useabstrax/abstrax/plugins/composer/internal/php"
	"github.com/useabstrax/abstrax/plugins/composer/sdk/abstrax"
)

func writePHP(t *testing.T, dir, name, version string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"-r\" ]; then echo '" + version + "'; exit 0; fi\n" +
		"if [ \"$1\" = \"-m\" ]; then printf 'json\\nmbstring\\nxml\\nzip\\ncurl\\n'; exit 0; fi\n" +
		"echo php \"$@\"\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestResolveFlagBeatsEnvAndConfig(t *testing.T) {
	dir := t.TempDir()
	writePHP(t, dir, "php8.2", "8.2.0")
	writePHP(t, dir, "php8.4", "8.4.0")
	writePHP(t, dir, "php", "8.5.0")
	t.Setenv("PATH", dir)
	t.Setenv(php.EnvPHP, "php8.4")

	res, err := php.Resolve(php.Options{Flag: "php8.2", ConfigPHP: "php8.4"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != php.SourceFlag {
		t.Fatalf("source = %s", res.Source)
	}
	if filepath.Base(res.Binary) != "php8.2" {
		t.Fatalf("binary = %s", res.Binary)
	}
}

func TestResolveEnvBeatsConfig(t *testing.T) {
	dir := t.TempDir()
	writePHP(t, dir, "php8.2", "8.2.0")
	writePHP(t, dir, "php8.4", "8.4.0")
	t.Setenv("PATH", dir)
	t.Setenv(php.EnvPHP, "php8.2")

	res, err := php.Resolve(php.Options{ConfigPHP: "php8.4"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != php.SourceEnv {
		t.Fatalf("source = %s", res.Source)
	}
}

func TestResolveProjectBeatsConfig(t *testing.T) {
	dir := t.TempDir()
	writePHP(t, dir, "php8.5", "8.5.0")
	writePHP(t, dir, "php8.2", "8.2.0")
	t.Setenv("PATH", dir)
	t.Setenv(php.EnvPHP, "")

	proj := &abstrax.Project{Runtime: abstrax.ProjectRuntime{Type: "php", Version: "8.5"}}
	res, err := php.Resolve(php.Options{ConfigPHP: "php8.2", Project: proj})
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != php.SourceProject {
		t.Fatalf("source = %s", res.Source)
	}
	if filepath.Base(res.Binary) != "php8.5" {
		t.Fatalf("binary = %s", res.Binary)
	}
}

func TestResolveStaticProjectFallsBackToConfig(t *testing.T) {
	dir := t.TempDir()
	writePHP(t, dir, "php8.2", "8.2.0")
	t.Setenv("PATH", dir)
	t.Setenv(php.EnvPHP, "")

	proj := &abstrax.Project{Runtime: abstrax.ProjectRuntime{Type: "static"}}
	res, err := php.Resolve(php.Options{ConfigPHP: "php8.2", Project: proj})
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != php.SourceConfig {
		t.Fatalf("source = %s", res.Source)
	}
}

func TestResolveDefaultPHP(t *testing.T) {
	dir := t.TempDir()
	writePHP(t, dir, "php", "8.5.0")
	t.Setenv("PATH", dir)
	t.Setenv(php.EnvPHP, "")

	res, err := php.Resolve(php.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Source != php.SourceDefault {
		t.Fatalf("source = %s", res.Source)
	}
}

func TestResolveMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv(php.EnvPHP, "")
	_, err := php.Resolve(php.Options{Flag: "php8.1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidate(t *testing.T) {
	dir := t.TempDir()
	writePHP(t, dir, "php8.2", "8.2.18")
	t.Setenv("PATH", dir)
	res, err := php.Validate(context.Background(), "php8.2")
	if err != nil {
		t.Fatal(err)
	}
	if res.Version != "8.2.18" {
		t.Fatalf("version = %q", res.Version)
	}
}
