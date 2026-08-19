package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
)

// File is a Composer auth.json document. Only the fields this plugin manages
// are modelled; unknown keys are not round-tripped.
type File struct {
	GithubOAuth map[string]string    `json:"github-oauth,omitempty"`
	GitlabOAuth map[string]string    `json:"gitlab-oauth,omitempty"`
	HTTPBasic   map[string]HTTPBasic `json:"http-basic,omitempty"`
}

// HTTPBasic is a username/password pair for a Composer repository host.
type HTTPBasic struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// PathForUser returns ~/.config/composer/auth.json for username.
func PathForUser(username string) (string, error) {
	var home string
	if strings.TrimSpace(username) == "" {
		u, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("resolving current user: %w", err)
		}
		home = u.HomeDir
	} else {
		u, err := user.Lookup(username)
		if err != nil {
			return "", fmt.Errorf("looking up user %q: %w", username, err)
		}
		home = u.HomeDir
	}
	if home == "" {
		return "", fmt.Errorf("user %q has no home directory", username)
	}
	return filepath.Join(home, ".config", "composer", "auth.json"), nil
}

// Load reads auth.json. Missing files yield an empty File.
func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return File{}, nil
		}
		return File{}, err
	}
	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		return File{}, fmt.Errorf("invalid Composer auth.json: %w", err)
	}
	return file, nil
}

// Save writes auth.json with mode 0600.
func Save(path string, file File) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// SetGithubToken stores a GitHub OAuth token for github.com.
func (f *File) SetGithubToken(token string) {
	token = strings.TrimSpace(token)
	if f.GithubOAuth == nil {
		f.GithubOAuth = map[string]string{}
	}
	f.GithubOAuth["github.com"] = token
}

// ClearGithub removes GitHub OAuth tokens.
func (f *File) ClearGithub() {
	f.GithubOAuth = nil
}

// SetHTTPBasic stores credentials for host.
func (f *File) SetHTTPBasic(host, username, password string) {
	host = strings.TrimSpace(host)
	if f.HTTPBasic == nil {
		f.HTTPBasic = map[string]HTTPBasic{}
	}
	f.HTTPBasic[host] = HTTPBasic{Username: username, Password: password}
}

// ClearHTTPBasic removes credentials for host, or all hosts when host is empty.
func (f *File) ClearHTTPBasic(host string) {
	host = strings.TrimSpace(host)
	if host == "" {
		f.HTTPBasic = nil
		return
	}
	delete(f.HTTPBasic, host)
	if len(f.HTTPBasic) == 0 {
		f.HTTPBasic = nil
	}
}

// HasGithub reports whether a GitHub token is stored.
func (f File) HasGithub() bool {
	return strings.TrimSpace(f.GithubOAuth["github.com"]) != ""
}

// HTTPBasicHosts returns configured repository hosts.
func (f File) HTTPBasicHosts() []string {
	hosts := make([]string, 0, len(f.HTTPBasic))
	for host := range f.HTTPBasic {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	return hosts
}

// Redacted returns a copy safe to print. Token values are replaced.
func (f File) Redacted() File {
	out := File{}
	if f.HasGithub() {
		out.GithubOAuth = map[string]string{"github.com": "********"}
	}
	if len(f.HTTPBasic) > 0 {
		out.HTTPBasic = map[string]HTTPBasic{}
		for host, cred := range f.HTTPBasic {
			out.HTTPBasic[host] = HTTPBasic{
				Username: cred.Username,
				Password: "********",
			}
		}
	}
	return out
}
