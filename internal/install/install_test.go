package install_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/useabstrax/abstrax/plugins/composer/internal/install"
	"github.com/useabstrax/abstrax/plugins/composer/internal/layout"
)

func TestLatestStableVerifiesChecksum(t *testing.T) {
	phar := []byte("fake-composer-phar")
	sum := sha256.Sum256(phar)
	hash := hex.EncodeToString(sum[:])

	mux := http.NewServeMux()
	mux.HandleFunc("/versions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"stable":[{"path":"/download/2.8.0/composer.phar","version":"2.8.0"}]}`)
	})
	mux.HandleFunc("/download/2.8.0/composer.phar.sha256sum", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  composer.phar\n", hash)
	})
	mux.HandleFunc("/download/2.8.0/composer.phar", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(phar)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	dl := &install.Downloader{BaseURL: srv.URL, HTTPClient: srv.Client()}
	rel, err := dl.LatestStable(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rel.Version != "2.8.0" {
		t.Fatalf("version = %s", rel.Version)
	}
	if !bytes.Equal(rel.Phar, phar) {
		t.Fatal("phar mismatch")
	}
}

func TestLatestStableChecksumMismatch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/versions", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"stable":[{"path":"/download/2.8.0/composer.phar","version":"2.8.0"}]}`)
	})
	mux.HandleFunc("/download/2.8.0/composer.phar.sha256sum", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  composer.phar\n")
	})
	mux.HandleFunc("/download/2.8.0/composer.phar", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("phar"))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	dl := &install.Downloader{BaseURL: srv.URL, HTTPClient: srv.Client()}
	_, err := dl.LatestStable(context.Background())
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch, got %v", err)
	}
}

func TestWritePharAndWrapper(t *testing.T) {
	dir := t.TempDir()
	paths := layout.Paths{
		LibDir:      filepath.Join(dir, "lib"),
		PharName:    "composer.phar",
		WrapperPath: filepath.Join(dir, "bin", "composer"),
	}
	if err := install.WritePhar(paths, []byte("phar")); err != nil {
		t.Fatal(err)
	}
	if err := install.WriteWrapper(paths, "/usr/bin/php8.2"); err != nil {
		t.Fatal(err)
	}
	if !install.IsManagedWrapper(paths.WrapperPath) {
		t.Fatal("expected managed wrapper")
	}
	body, err := os.ReadFile(paths.WrapperPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "/usr/bin/php8.2") {
		t.Fatalf("wrapper = %s", body)
	}
	if !strings.Contains(string(body), paths.PharPath()) {
		t.Fatalf("wrapper missing phar path: %s", body)
	}
}

func TestIsManagedWrapperFalse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "composer")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexec php /usr/bin/composer.phar\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if install.IsManagedWrapper(path) {
		t.Fatal("distro wrapper should not be managed")
	}
}
