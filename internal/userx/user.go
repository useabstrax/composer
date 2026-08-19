package userx

import (
	"fmt"
	"os"
	"os/user"
	"strings"
)

// RequireRoot is the root check used by mutating commands. Tests may replace it.
var RequireRoot = requireRoot

func requireRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this command requires root (try sudo)")
	}
	return nil
}

// Lookup returns the named user.
func Lookup(username string) (*user.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("user is required")
	}
	u, err := user.Lookup(username)
	if err != nil {
		return nil, fmt.Errorf("looking up user %q: %w", username, err)
	}
	return u, nil
}

// CurrentUsername returns the process user name.
func CurrentUsername() string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return ""
}

// EffectiveRunUser chooses who Composer should run as.
// When invoked with sudo and no explicit user, SUDO_USER is preferred over root.
func EffectiveRunUser(explicit string, allowRoot bool) (string, error) {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" {
		if explicit == "root" && !allowRoot {
			return "", fmt.Errorf("refusing to run Composer as root; pass --user or --allow-root")
		}
		return explicit, nil
	}
	if os.Geteuid() != 0 {
		return CurrentUsername(), nil
	}
	if sudoUser := strings.TrimSpace(os.Getenv("SUDO_USER")); sudoUser != "" && sudoUser != "root" {
		return sudoUser, nil
	}
	if allowRoot {
		return "root", nil
	}
	return "", fmt.Errorf("refusing to run Composer as root; pass --user or --allow-root")
}

// IsSharedOwner reports whether the user is a shared web account.
func IsSharedOwner(username string) bool {
	u := strings.ToLower(strings.TrimSpace(username))
	return u == "" || u == "www-data" || u == "nginx" || u == "apache"
}
