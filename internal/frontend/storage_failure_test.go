package frontend

import (
	"os"
	"path/filepath"
	"testing"

	"pinballfantasies/internal/settings"
)

// A state folder that cannot be written must not end the game.
func TestSettingsWriteFailureStillQuits(t *testing.T) {
	m, err := New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, nil, 0600); err != nil {
		t.Fatal(err)
	}
	m.SettingsStore = &settings.Store{Directory: blocked}
	m.Mode = Selector
	if err := m.Update(Input{Keys: []Key{Escape}}); err != nil || m.Mode != Quit {
		t.Fatalf("selector quit: %v %v", err, m.Mode)
	}
	if err := m.Update(Input{Close: true}); err != nil || m.Mode != Quit {
		t.Fatalf("close: %v %v", err, m.Mode)
	}
}
