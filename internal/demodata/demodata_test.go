package demodata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pinballfantasies/internal/frontend"
)

func TestBundledDemoLoads(t *testing.T) {
	dir, cleanup, err := Extract()
	if err != nil {
		t.Fatal(err)
	}
	r, err := frontend.Load(dir, frontend.FileStore{Directory: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if r.ProfileID != frontend.DemoProfileID || len(r.Model.ClosingText) == 0 {
		t.Fatalf("profile %s, closing text %q", r.ProfileID, r.Model.ClosingText)
	}
	cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("demo folder left behind: %v", err)
	}
}

// Copies from CDs and archives often arrive with lower-case names.
func TestLowerCaseDemoLoads(t *testing.T) {
	dir, cleanup, err := Extract()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, f := range Files {
		if err := os.Rename(filepath.Join(dir, f.Name), filepath.Join(dir, strings.ToLower(f.Name))); err != nil {
			t.Fatal(err)
		}
	}
	r, err := frontend.Load(dir, frontend.FileStore{Directory: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if r.ProfileID != frontend.DemoProfileID || len(r.Model.ClosingText) == 0 {
		t.Fatalf("profile %s, closing text %q", r.ProfileID, r.Model.ClosingText)
	}
}
