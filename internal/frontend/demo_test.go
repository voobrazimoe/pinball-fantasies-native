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

func TestPrivateDemoRuntimeLifecycle(t *testing.T) {
	dir := privateDemoDir(t)
	state := t.TempDir()
	r, err := Load(dir, FileStore{Directory: state})
	if err != nil {
		t.Fatal(err)
	}
	if r.ProfileID != DemoProfileID || r.Model.Mode != Playing || r.Model.Selected != 1 || r.Model.Hz() != 71 {
		t.Fatal(r.ProfileID, r.Model.Mode)
	}
	for i := 0; i < 200; i++ {
		if err := r.Update(Input{}); err != nil {
			t.Fatal(err)
		}
	}
	if demoTimer(t, r) != 200 || r.Frame() == nil {
		t.Fatal(demoTimer(t, r))
	}
	// Pause: no electronics, no audio, timer and session unchanged.
	if err := r.Update(Input{Keys: []Key{P}}); err != nil || r.Model.Mode != Paused {
		t.Fatal(err, r.Model.Mode)
	}
	for i := 0; i < 500; i++ {
		if err := r.Update(Input{}); err != nil || r.PCM != nil {
			t.Fatal(err)
		}
	}
	if demoTimer(t, r) != 200 {
		t.Fatal("paused calculation counted")
	}
	if err := r.Update(Input{Keys: []Key{Space}}); err != nil || r.Model.Mode != Playing {
		t.Fatal(r.Model.Mode)
	}
	// Esc asks; any other key resumes; Y quits.
	r.Update(Input{Keys: []Key{Escape}})
	if r.Model.Mode != QuitQuestion {
		t.Fatal(r.Model.Mode)
	}
	r.Update(Input{Keys: []Key{Key(49)}}) // N
	if r.Model.Mode != Playing {
		t.Fatal(r.Model.Mode)
	}
	r.Update(Input{Keys: []Key{Escape}})
	r.Update(Input{Keys: []Key{Key(21)}}) // Y
	if r.Model.Mode != Quit || r.Model.End != Aborted {
		t.Fatal(r.Model.Mode)
	}
	// Volatile scores/settings: nothing is written beside the user's data.
	if entries, _ := os.ReadDir(state); len(entries) != 0 {
		t.Fatal("demo session wrote state", entries)
	}
}

func TestPrivateDemoRuntimeEndsAtQuit(t *testing.T) {
	r, err := Load(privateDemoDir(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 40000 && r.Model.Mode != Quit; i++ {
		if err := r.Update(Input{}); err != nil {
			t.Fatal(err)
		}
	}
	if r.Model.Mode != Quit || r.Model.End != ProgramQuit || demoTimer(t, r) != 37047 {
		t.Fatal(r.Model.Mode, demoTimer(t, r))
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
