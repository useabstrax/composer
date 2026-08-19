package auth_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/useabstrax/abstrax/plugins/composer/internal/auth"
)

func TestSaveLoadRedact(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.json")
	file := auth.File{}
	file.SetGithubToken("ghp_secret")
	file.SetHTTPBasic("repo.packagist.com", "token", "secret")
	if err := auth.Save(path, file); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %s", st.Mode())
	}

	got, err := auth.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.HasGithub() {
		t.Fatal("expected github token")
	}
	if got.GithubOAuth["github.com"] != "ghp_secret" {
		t.Fatal("token not stored")
	}

	redacted := got.Redacted()
	if redacted.GithubOAuth["github.com"] != "********" {
		t.Fatalf("redacted github = %q", redacted.GithubOAuth["github.com"])
	}
	if redacted.HTTPBasic["repo.packagist.com"].Password != "********" {
		t.Fatal("password not redacted")
	}
	if redacted.HTTPBasic["repo.packagist.com"].Username != "token" {
		t.Fatal("username should remain")
	}
}

func TestClearGithub(t *testing.T) {
	file := auth.File{}
	file.SetGithubToken("x")
	file.ClearGithub()
	if file.HasGithub() {
		t.Fatal("expected cleared")
	}
}

func TestLoadMissing(t *testing.T) {
	got, err := auth.Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got.HasGithub() {
		t.Fatal("expected empty file")
	}
}
