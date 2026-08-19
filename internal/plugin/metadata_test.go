package plugin_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/useabstrax/abstrax/plugins/composer/internal/plugin"
)

func TestDefaultMetadataJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := plugin.WriteMetadata(&buf, plugin.DefaultMetadata()); err != nil {
		t.Fatalf("WriteMetadata: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	required := []string{
		"protocol_version", "name", "display_name", "description",
		"version", "requires_abstrax", "commands",
	}
	for _, key := range required {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing required field %q", key)
		}
	}
	if raw["name"] != plugin.PluginName {
		t.Errorf("name = %v", raw["name"])
	}
	if raw["protocol_version"].(float64) != float64(plugin.ProtocolVersion) {
		t.Errorf("protocol_version = %v", raw["protocol_version"])
	}
}

func TestDefaultMetadataCommands(t *testing.T) {
	meta := plugin.DefaultMetadata()
	if meta.Name != "composer" {
		t.Fatalf("name = %q", meta.Name)
	}
	names := map[string]bool{}
	actions := map[string]string{}
	for _, c := range meta.Commands {
		names[c.Name] = true
		actions[c.Name] = c.Action
	}
	for _, want := range []string{"setup", "self-update", "remove", "status", "configure", "run", "diagnose", "auth", "version"} {
		if !names[want] {
			t.Fatalf("missing command %q", want)
		}
	}
	if actions["self-update"] != "plugin.composer.self_update" {
		t.Fatalf("self-update action = %q", actions["self-update"])
	}
}

func TestWriteMetadataNoTrailingGarbage(t *testing.T) {
	var buf bytes.Buffer
	if err := plugin.WriteMetadata(&buf, plugin.DefaultMetadata()); err != nil {
		t.Fatal(err)
	}
	if !json.Valid(buf.Bytes()) {
		t.Fatalf("output is not valid JSON: %s", buf.String())
	}
}

func TestReleaseManifestExampleMatchesFixtureShape(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	examplePath := filepath.Join(filepath.Dir(file), "..", "..", "plugin-manifest.example.json")
	example, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(example, &got); err != nil {
		t.Fatal(err)
	}
	required := []string{"name", "version", "protocol_version", "requires_abstrax", "channel", "platforms"}
	for _, key := range required {
		if _, ok := got[key]; !ok {
			t.Fatalf("example manifest missing %q", key)
		}
	}
	if got["name"] != "composer" {
		t.Fatalf("name = %v", got["name"])
	}
}

func TestIsRunningAsPlugin(t *testing.T) {
	t.Setenv("ABSTRAX_PLUGIN", "1")
	if !plugin.IsRunningAsPlugin() {
		t.Fatal("expected true when ABSTRAX_PLUGIN=1")
	}
	t.Setenv("ABSTRAX_PLUGIN", "")
	if plugin.IsRunningAsPlugin() {
		t.Fatal("expected false when ABSTRAX_PLUGIN is empty")
	}
}

func TestHostVersion(t *testing.T) {
	t.Setenv("ABSTRAX_VERSION", "1.2.3")
	if got := plugin.HostVersion(); got != "1.2.3" {
		t.Fatalf("HostVersion() = %q", got)
	}
}

func TestHostBinary(t *testing.T) {
	t.Setenv("ABSTRAX_BINARY", "/usr/bin/abstrax")
	if got := plugin.HostBinary(); got != "/usr/bin/abstrax" {
		t.Fatalf("HostBinary() = %q", got)
	}
}
