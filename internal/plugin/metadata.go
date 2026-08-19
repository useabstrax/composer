package plugin

import (
	"encoding/json"
	"io"
	"strings"
)

const (
	// ProtocolVersion is the supported plugin metadata protocol version.
	ProtocolVersion = 1

	PluginName      = "composer"
	DisplayName     = "Composer Plugin"
	Description     = "Install and manage Composer, and run it with the correct PHP binary"
	Homepage        = "https://plugins.useabstrax.com/plugins/composer"
	RequiresAbstrax = ">=0.1.0"
)

// MetadataCommand describes a subcommand exposed by the plugin.
type MetadataCommand struct {
	Name        string `json:"name"`
	Action      string `json:"action,omitempty"`
	Description string `json:"description"`
}

// Metadata is the plugin metadata protocol v1 response.
type Metadata struct {
	ProtocolVersion int               `json:"protocol_version"`
	Name            string            `json:"name"`
	DisplayName     string            `json:"display_name"`
	Description     string            `json:"description"`
	Version         string            `json:"version"`
	RequiresAbstrax string            `json:"requires_abstrax"`
	Homepage        string            `json:"homepage,omitempty"`
	Commands        []MetadataCommand `json:"commands"`
}

// DefaultMetadata returns the plugin metadata for this build.
func DefaultMetadata() Metadata {
	return Metadata{
		ProtocolVersion: ProtocolVersion,
		Name:            PluginName,
		DisplayName:     DisplayName,
		Description:     Description,
		Version:         Version,
		RequiresAbstrax: RequiresAbstrax,
		Homepage:        Homepage,
		Commands: []MetadataCommand{
			composerCommand("setup", "Download Composer, verify it, and install it globally"),
			composerCommand("self-update", "Update the installed Composer phar to the latest stable version"),
			composerCommand("remove", "Remove the Composer binary installed by this plugin"),
			composerCommand("status", "Show Composer install state and the resolved PHP binary"),
			composerCommand("configure", "Show or set the default PHP binary for Composer"),
			composerCommand("run", "Run Composer with the resolved PHP binary"),
			composerCommand("diagnose", "Check Composer, PHP, and common server prerequisites"),
			composerCommand("auth", "Show or update Composer authentication for a user"),
			composerCommand("version", "Display plugin version information"),
		},
	}
}

func composerCommand(name, description string) MetadataCommand {
	return MetadataCommand{
		Name:        name,
		Action:      "plugin." + PluginName + "." + strings.ReplaceAll(name, "-", "_"),
		Description: description,
	}
}

// WriteMetadata encodes metadata as indented JSON to w.
func WriteMetadata(w io.Writer, metadata Metadata) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(metadata)
}
