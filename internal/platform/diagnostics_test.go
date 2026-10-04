//go:build linux || windows

package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDiagnosticsDoNotCreateFiles(t *testing.T) {
	old := GUIMode
	GUIMode = "1"
	defer func() { GUIMode = old }()
	t.Setenv("PF_DIAGNOSTICS", "")
	directory := filepath.Join(t.TempDir(), "untouched")
	if err := InitDiagnostics(directory); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(directory); !os.IsNotExist(err) {
		t.Fatalf("diagnostics touched state directory: %v", err)
	}
}
