package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// InstallFakeAbstrax writes a fake abstrax executable and returns its path.
func InstallFakeAbstrax(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "abstrax")
	script := `#!/bin/sh
set -eu

if [ "$1" = "project" ] && [ "$2" = "inspect" ]; then
  name="$3"
  json="$4"
  if [ "$json" != "--json" ]; then
    echo "expected --json" >&2
    exit 1
  fi
  case "$name" in
    exists)
      cat <<'EOF'
{
  "api_version": "v1",
  "project": {
    "name": "example",
    "path": "/var/www/example.com",
    "user": "example",
    "runtime": {"type": "php", "version": "8.5"},
    "domains": ["example.com"],
    "services": [{"name": "example-worker", "type": "worker"}]
  }
}
EOF
      exit 0
      ;;
    static)
      cat <<'EOF'
{
  "api_version": "v1",
  "project": {
    "name": "static",
    "path": "/var/www/static",
    "user": "www-data",
    "runtime": {"type": "static", "version": ""},
    "domains": ["static.example"],
    "services": []
  }
}
EOF
      exit 0
      ;;
    missing)
      echo 'project "missing" not found' >&2
      exit 1
      ;;
    badjson)
      echo 'not json'
      exit 0
      ;;
    oldapi)
      cat <<'EOF'
{"api_version": "v0", "project": {"name": "old"}}
EOF
      exit 0
      ;;
    *)
      echo "unknown project $name" >&2
      exit 1
      ;;
  esac
fi

echo "unsupported invocation: $*" >&2
exit 1
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake abstrax: %v", err)
	}
	return path
}

// WithEnvBinary sets ABSTRAX_BINARY for the duration of the test.
func WithEnvBinary(t *testing.T, path string) {
	t.Helper()
	t.Setenv("ABSTRAX_BINARY", path)
}

// AssertContains fails if s does not contain substr.
func AssertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q to contain %q", s, substr)
	}
}
