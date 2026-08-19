package abstrax_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/useabstrax/abstrax/plugins/composer/sdk/abstrax"
	"github.com/useabstrax/abstrax/plugins/composer/sdk/abstrax/testutil"
)

func TestProjectSuccess(t *testing.T) {
	fake := testutil.InstallFakeAbstrax(t)
	client, err := abstrax.NewWithBinary(fake)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Project(context.Background(), "exists")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Project.Name != "example" {
		t.Fatalf("name = %q", resp.Project.Name)
	}
}

func TestProjectNotFound(t *testing.T) {
	fake := testutil.InstallFakeAbstrax(t)
	client, err := abstrax.NewWithBinary(fake)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Project(context.Background(), "missing")
	if !errors.Is(err, abstrax.ErrProjectNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestProjectMalformedJSON(t *testing.T) {
	fake := testutil.InstallFakeAbstrax(t)
	client, err := abstrax.NewWithBinary(fake)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Project(context.Background(), "badjson")
	if !errors.Is(err, abstrax.ErrMalformedJSON) {
		t.Fatalf("got %v", err)
	}
}

func TestProjectUnsupportedAPIVersion(t *testing.T) {
	fake := testutil.InstallFakeAbstrax(t)
	client, err := abstrax.NewWithBinary(fake)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Project(context.Background(), "oldapi")
	if !errors.Is(err, abstrax.ErrUnsupportedAPIVersion) {
		t.Fatalf("got %v", err)
	}
}

func TestProjectResponseUnmarshal(t *testing.T) {
	raw := `{
		"api_version": "v1",
		"project": {
			"name": "example",
			"path": "/var/www/example.com",
			"user": "example",
			"runtime": {"type": "php", "version": "8.5"},
			"domains": ["example.com"],
			"services": []
		}
	}`
	var resp abstrax.ProjectResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Project.Runtime.Version != "8.5" {
		t.Fatalf("version = %q", resp.Project.Runtime.Version)
	}
}

func TestNewPrefersABSTRAXBinary(t *testing.T) {
	fake := testutil.InstallFakeAbstrax(t)
	testutil.WithEnvBinary(t, fake)
	client, err := abstrax.New()
	if err != nil {
		t.Fatal(err)
	}
	if client.Binary != fake {
		t.Fatalf("binary = %q", client.Binary)
	}
}

func TestNewBinaryNotFound(t *testing.T) {
	t.Setenv("ABSTRAX_BINARY", "")
	t.Setenv("PATH", t.TempDir())
	_, err := abstrax.New()
	if !errors.Is(err, abstrax.ErrBinaryNotFound) {
		t.Fatalf("got %v", err)
	}
}
