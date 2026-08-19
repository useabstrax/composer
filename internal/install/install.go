package install

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/useabstrax/abstrax/plugins/composer/internal/layout"
	"github.com/useabstrax/abstrax/plugins/composer/internal/plugin"
)

const (
	defaultBaseURL = "https://getcomposer.org"
	maxPharBytes   = 50 << 20
)

// HTTPDoer performs HTTP requests. Tests inject a fake client.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Downloader fetches and verifies the Composer phar.
type Downloader struct {
	BaseURL    string
	HTTPClient HTTPDoer
	UserAgent  string
}

type versionsResponse struct {
	Stable []versionEntry `json:"stable"`
}

type versionEntry struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

// Release is a verified Composer download.
type Release struct {
	Version  string
	Phar     []byte
	SHA256   string
	Download string
}

func (d *Downloader) client() HTTPDoer {
	if d.HTTPClient != nil {
		return d.HTTPClient
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (d *Downloader) base() string {
	if strings.TrimSpace(d.BaseURL) != "" {
		return strings.TrimRight(d.BaseURL, "/")
	}
	return defaultBaseURL
}

func (d *Downloader) userAgent() string {
	if d.UserAgent != "" {
		return d.UserAgent
	}
	return "Abstrax-Composer/" + plugin.Version + " (+https://useabstrax.com)"
}

func (d *Downloader) get(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", d.userAgent())
	resp, err := d.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("GET %s: response too large", url)
	}
	return body, nil
}

// LatestStable downloads and verifies the current stable Composer phar.
func (d *Downloader) LatestStable(ctx context.Context) (*Release, error) {
	raw, err := d.get(ctx, d.base()+"/versions", 1<<20)
	if err != nil {
		return nil, fmt.Errorf("fetching Composer versions: %w", err)
	}
	var versions versionsResponse
	if err := json.Unmarshal(raw, &versions); err != nil {
		return nil, fmt.Errorf("parsing Composer versions: %w", err)
	}
	if len(versions.Stable) == 0 || versions.Stable[0].Path == "" {
		return nil, fmt.Errorf("Composer versions response had no stable release")
	}
	entry := versions.Stable[0]
	pharURL := d.base() + entry.Path
	sumURL := pharURL + ".sha256sum"

	sumBody, err := d.get(ctx, sumURL, 4096)
	if err != nil {
		return nil, fmt.Errorf("fetching Composer checksum: %w", err)
	}
	want, err := parseSHA256Sum(string(sumBody))
	if err != nil {
		return nil, err
	}

	phar, err := d.get(ctx, pharURL, maxPharBytes)
	if err != nil {
		return nil, fmt.Errorf("downloading Composer: %w", err)
	}
	sum := sha256.Sum256(phar)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, want) {
		return nil, fmt.Errorf("Composer checksum mismatch (got %s, want %s)", got, want)
	}

	return &Release{
		Version:  entry.Version,
		Phar:     phar,
		SHA256:   want,
		Download: pharURL,
	}, nil
}

func parseSHA256Sum(body string) (string, error) {
	line := strings.TrimSpace(strings.Split(body, "\n")[0])
	if line == "" {
		return "", fmt.Errorf("empty Composer checksum file")
	}
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", fmt.Errorf("invalid Composer checksum file")
	}
	hash := strings.ToLower(fields[0])
	if len(hash) != 64 {
		return "", fmt.Errorf("invalid Composer checksum %q", fields[0])
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return "", fmt.Errorf("invalid Composer checksum %q", fields[0])
	}
	return hash, nil
}

// WritePhar writes the phar into the plugin lib directory.
func WritePhar(paths layout.Paths, phar []byte) error {
	if err := os.MkdirAll(paths.LibDir, 0o755); err != nil {
		return err
	}
	path := paths.PharPath()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, phar, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// WriteWrapper installs /usr/local/bin/composer pointing at php + phar.
func WriteWrapper(paths layout.Paths, phpBinary string) error {
	script := WrapperScript(phpBinary, paths.PharPath())
	dir := filepath.Dir(paths.WrapperPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp := paths.WrapperPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(script), 0o755); err != nil {
		return err
	}
	return os.Rename(tmp, paths.WrapperPath)
}

// WrapperScript returns the shell wrapper contents.
func WrapperScript(phpBinary, pharPath string) string {
	return fmt.Sprintf("#!/bin/sh\n# %s. Do not edit.\nexec %s %s \"$@\"\n",
		layout.WrapperMarker, shellQuote(phpBinary), shellQuote(pharPath))
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// IsManagedWrapper reports whether path is an Abstrax-managed Composer wrapper.
func IsManagedWrapper(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), layout.WrapperMarker)
}

// WrapperExists reports whether the wrapper path exists.
func WrapperExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
