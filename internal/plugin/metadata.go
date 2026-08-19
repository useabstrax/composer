package plugin

import (
	"encoding/json"
	"io"
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
			{Name: "setup", Description: "Download Composer, verify it, and install it globally"},
			{Name: "self-update", Description: "Update the installed Composer phar to the latest stable version"},
			{Name: "remove", Description: "Remove the Composer binary installed by this plugin"},
			{Name: "status", Description: "Show Composer install state and the resolved PHP binary"},
			{Name: "configure", Description: "Show or set the default PHP binary for Composer"},
			{Name: "run", Description: "Run Composer with the resolved PHP binary"},
			{Name: "diagnose", Description: "Check Composer, PHP, and common server prerequisites"},
			{Name: "auth", Description: "Show or update Composer authentication for a user"},
			{Name: "version", Description: "Display plugin version information"},
		},
	}
}

// WriteMetadata encodes metadata as indented JSON to w.
func WriteMetadata(w io.Writer, metadata Metadata) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(metadata)
}
