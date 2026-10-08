//go:build demodev

package frontend

import (
	"os"
	"path/filepath"
	"pinballfantasies/internal/gameplay"
	"strings"
	"testing"
)

func TestExperimentalLifecycleDiagnosticAndClose(t *testing.T) {
	demo, canonical := os.Getenv("PF_10MIN_DEMO_DATA"), os.Getenv("PF_RUNTIME_DATA")
	if demo == "" || canonical == "" {
		t.Skip("owner-local inputs NOT AVAILABLE")
	}
	log := filepath.Join(t.TempDir(), "demo.log")
	r, err := LoadExperimentalDemo(demo, canonical, log)
	if err != nil {
		t.Fatal(err)
	}
	s := r.session.(*experimentalSession)
	update := func(in Input) {
		t.Helper()
		if err := r.Update(in); err != nil {
			t.Fatal(err)
		}
	}
	update(Input{})
	n := s.game.Calculations()
	update(Input{Keys: []Key{P}})
	for i := 0; i < 10; i++ {
		update(Input{})
	}
	if s.game.Calculations() != n || len(r.PCM) != 0 {
		t.Fatal("paused physics/audio advanced")
	}
	update(Input{Keys: []Key{P}})
	if s.game.Calculations() != n+1 {
		t.Fatal("resume tick")
	}
	if err := r.FocusLost(false); err != nil {
		t.Fatal(err)
	}
	update(Input{})
	if s.game.Calculations() != n+1 {
		t.Fatal("focus pause tick")
	}
	update(Input{Keys: []Key{P}})
	update(Input{Gameplay: gameplay.Controls{Tilt: true}})
	if r.Model.Mode != Paused || !strings.Contains(r.Diagnostic(), "UNSUPPORTED_DEMO_TRANSITION") {
		t.Fatal("no safe diagnostic")
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"producer=", "consumer=", "calculation=", "phase=keyboard admission", "state/guard=", "reason="} {
		if !strings.Contains(string(data), key) {
			t.Fatal("diagnostic missing", key)
		}
	}
	n = s.game.Calculations()
	update(Input{Keys: []Key{P}})
	update(Input{})
	if s.game.Calculations() != n {
		t.Fatal("resumed failed gameplay")
	}
	update(Input{Keys: []Key{Escape}})
	if r.Model.Mode != Quit {
		t.Fatal("escape did not close")
	}
}
