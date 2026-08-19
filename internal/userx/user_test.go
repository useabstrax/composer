package userx_test

import (
	"os"
	"testing"

	"github.com/useabstrax/abstrax/plugins/composer/internal/userx"
)

func TestEffectiveRunUserExplicit(t *testing.T) {
	got, err := userx.EffectiveRunUser("deploy", false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "deploy" {
		t.Fatalf("got %q", got)
	}
}

func TestEffectiveRunUserRejectsRoot(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("not root")
	}
	_, err := userx.EffectiveRunUser("", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEffectiveRunUserAllowRoot(t *testing.T) {
	got, err := userx.EffectiveRunUser("root", true)
	if err != nil {
		t.Fatal(err)
	}
	if got != "root" {
		t.Fatalf("got %q", got)
	}
}

func TestIsSharedOwner(t *testing.T) {
	if !userx.IsSharedOwner("www-data") {
		t.Fatal("www-data should be shared")
	}
	if userx.IsSharedOwner("deploy") {
		t.Fatal("deploy should not be shared")
	}
}
