package frontend

import (
	"os"
	"path/filepath"
	"testing"

	"pinballfantasies/internal/partyland"
)

func privateDemoDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("PF_10MIN_DEMO_DATA")
	if dir == "" {
		t.Skip("set PF_10MIN_DEMO_DATA to the private 10-minute demo")
	}
	// Copy only the five runtime roles, so the launcher/drivers are irrelevant.
	out := t.TempDir()
	for _, name := range DemoNamesRequired {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(out, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return out
}

func demoTimer(t *testing.T, r *Runtime) uint16 {
	t.Helper()
	c, _, ok := r.Model.Session.(*partyland.Game).Demo()
	if !ok {
		t.Fatal("not a demo session")
	}
	return c
}

func demoKeys(t *testing.T, r *Runtime, keys ...Key) {
	t.Helper()
	for _, k := range keys {
		if err := r.Update(Input{Keys: []Key{k}}); err != nil {
			t.Fatal(err)
		}
	}
}
func demoFrames(t *testing.T, r *Runtime, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if err := r.Update(Input{}); err != nil {
			t.Fatal(err)
		}
		if i%97 == 0 && r.Frame() == nil {
			t.Fatal("no frame")
		}
	}
}

// INTRO -> selector -> options -> Party Land, with pause and quit question.
func TestPrivateDemoRuntimeLifecycle(t *testing.T) {
	dir := privateDemoDir(t)
	state := t.TempDir()
	r, err := Load(dir, FileStore{Directory: state})
	if err != nil {
		t.Fatal(err)
	}
	m := r.Model
	if r.ProfileID != DemoProfileID || !m.Demo || m.Mode != Startup || r.View == nil || len(r.View.Art.TextPages) != 2 {
		t.Fatal(r.ProfileID, m.Mode)
	}
	if m.Factories[1] != nil || m.Factories[2] != nil || m.Factories[3] != nil {
		t.Fatal("unavailable tables must have no factory")
	}
	demoFrames(t, r, 120)
	demoKeys(t, r, Space)
	if m.Mode != Selector {
		t.Fatal(m.Mode)
	}
	// The demo sidebar names only Party Land, options and quit.
	if r.View.Art.SidebarInfo[12:24] != "F1 -   PARTY" {
		t.Fatalf("%q", r.View.Art.SidebarInfo[:36])
	}
	demoKeys(t, r, F2, F3, F4)
	if m.Mode != Selector || m.Selected != 1 || m.Session != nil {
		t.Fatal("unavailable table selected", m.Mode, m.Selected)
	}
	// Selector text cycles the demo's own two SHOWTEXT pages.
	seen := map[int]bool{}
	for i := 0; i < 6000 && len(seen) < 2; i++ {
		demoFrames(t, r, 1)
		if m.Mode == SelectorText {
			seen[m.TextPage] = true
			if p := r.View.textPage(m); len(p) != 12 {
				t.Fatal(p)
			}
		}
	}
	if !seen[1] || !seen[2] {
		t.Fatal("demo text pages not reached", seen)
	}
	for m.Mode != Selector {
		demoFrames(t, r, 1)
	}
	demoKeys(t, r, F5)
	if m.Mode != Options {
		t.Fatal(m.Mode)
	}
	demoFrames(t, r, 60)
	demoKeys(t, r, Escape)
	demoFrames(t, r, 60)
	if m.Mode != Selector {
		t.Fatal("options did not return", m.Mode)
	}
	demoKeys(t, r, F1)
	if m.Mode != TableAttract || m.Selected != 1 {
		t.Fatal(m.Mode)
	}
	demoFrames(t, r, 50)
	demoKeys(t, r, F4) // single player in the demo, whatever the start key
	if m.Mode != Playing {
		t.Fatal(m.Mode)
	}
	if g, ok := m.Session.(interface{ PlayerCount() int }); !ok || g.PlayerCount() != 1 {
		t.Fatal("demo started more than one player")
	}
	if c := demoTimer(t, r); c != 0 {
		t.Fatal("table attract counted", c)
	}
	demoFrames(t, r, 200)
	if demoTimer(t, r) != 200 {
		t.Fatal(demoTimer(t, r))
	}
	demoKeys(t, r, P)
	demoFrames(t, r, 300)
	if m.Mode != Paused || demoTimer(t, r) != 200 {
		t.Fatal("paused calculation counted", m.Mode, demoTimer(t, r))
	}
	demoKeys(t, r, Escape, Key(21)) // Y
	if m.Mode != Selector || m.Session != nil {
		t.Fatal(m.Mode)
	}
	// Volatile scores: nothing is written to the native score store.
	if entries, _ := os.ReadDir(state); len(entries) != 0 {
		t.Fatal("demo wrote scores", entries)
	}
}

// QUIT(0) ends TABLE1 and the launcher runs INTRO again, with a fresh table
// lifetime on the next entry.
func TestPrivateDemoRuntimeQuitRestartsIntro(t *testing.T) {
	r, err := Load(privateDemoDir(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	m := r.Model
	demoKeys(t, r, Space, F1, F1)
	if m.Mode != Playing {
		t.Fatal(m.Mode)
	}
	for i := 0; i < 40000 && m.Mode == Playing; i++ {
		if err := r.Update(Input{}); err != nil {
			t.Fatal(err)
		}
	}
	if m.Mode != Startup || m.Session != nil || m.End != Completed {
		t.Fatal(m.Mode)
	}
	demoFrames(t, r, 30)
	demoKeys(t, r, Space, F1, F1)
	if m.Mode != Playing || demoTimer(t, r) != 0 {
		t.Fatal("next table entry is not a fresh lifetime")
	}
}

func TestPrivateDemoFolderErrors(t *testing.T) {
	dir := privateDemoDir(t)
	// A folder with a partial full installation keeps its ordinary error.
	if err := os.WriteFile(filepath.Join(dir, "TABLE2.PRG"), []byte{0}, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir, nil); err == nil || !os.IsNotExist(err) {
		t.Fatal("partial full installation must report the missing file", err)
	}
	os.Remove(filepath.Join(dir, "TABLE2.PRG"))
	b, _ := os.ReadFile(filepath.Join(dir, "TABLE1.PRG"))
	b[0x1ba17+5] ^= 0xff // inside the reviewed expiry stream; unconsumed bytes are not identity
	os.WriteFile(filepath.Join(dir, "TABLE1.PRG"), b, 0600)
	if _, err := Load(dir, nil); err == nil {
		t.Fatal("altered demo accepted")
	}
}
