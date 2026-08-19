package layout

import "path/filepath"

const (
	// WrapperMarker is written into the Composer wrapper so later commands
	// can tell it apart from a distro or hand-installed binary.
	WrapperMarker = "Managed by abstrax composer"
)

// Paths locates plugin-managed files. Tests override the exported defaults.
type Paths struct {
	ConfigPath  string
	LibDir      string
	PharName    string
	WrapperPath string
}

// DefaultPaths is the production filesystem layout.
var DefaultPaths = Paths{
	ConfigPath:  "/etc/abstrax/composer.json",
	LibDir:      "/usr/local/lib/abstrax/composer",
	PharName:    "composer.phar",
	WrapperPath: "/usr/local/bin/composer",
}

// PharPath returns the installed composer.phar path.
func (p Paths) PharPath() string {
	name := p.PharName
	if name == "" {
		name = "composer.phar"
	}
	return filepath.Join(p.LibDir, name)
}
