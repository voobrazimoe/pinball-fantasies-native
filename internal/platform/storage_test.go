package platform

import (
	"os"
	"path/filepath"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/settings"
	"runtime"
	"strings"
	"testing"
)

func TestPortableRoots(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Portable Game Ж 日本")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	exe, app := filepath.Join(root, "pinballfantasies.exe"), filepath.Join(root, "Game.AppImage")
	for _, p := range []string{exe, app} {
		if err := os.WriteFile(p, nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		exe, app string
		linux    bool
	}{
		{exe, "", false}, {exe, "", true}, {"/nonexistent/mount/game", app, true}, {exe, app, false},
	} {
		got, err := portableRoot(tc.exe, tc.app, tc.linux)
		if err != nil || !sameDirectory(got, root) {
			t.Fatalf("root=%q err=%v", got, err)
		}
	}
	if _, err := portableRoot(exe, filepath.Join(root, "missing.AppImage"), true); err == nil {
		t.Fatal("invalid APPIMAGE must not silently use mounted exe")
	}
	link := filepath.Join(t.TempDir(), "linked-game")
	if err := os.Symlink(exe, link); err == nil && runtime.GOOS != "windows" {
		if got, err := portableRoot(link, "", true); err != nil || !sameDirectory(got, root) {
			t.Fatalf("symlink root %q %v", got, err)
		}
	}
}

func TestPortableWritableStorage(t *testing.T) {
	root := t.TempDir()
	original := filepath.Join(root, "PINBALL.CFG")
	seed := []byte("original untouched")
	os.WriteFile(original, seed, 0400)
	state := filepath.Join(root, "userdata", "Ж space")
	if got, err := writableDirectory(state); err != nil || !sameDirectory(got, state) {
		t.Fatalf("%q %v", got, err)
	}
	entries, _ := os.ReadDir(state)
	if len(entries) != 0 {
		t.Fatal("write probe retained files")
	}
	if b, _ := os.ReadFile(original); string(b) != string(seed) {
		t.Fatal("seed modified")
	}
	blocked := filepath.Join(root, "blocked")
	os.WriteFile(blocked, nil, 0600)
	if _, err := writableDirectory(filepath.Join(blocked, "userdata")); err == nil || !strings.Contains(err.Error(), "move the game to a writable directory") {
		t.Fatalf("actionable failure: %v", err)
	}
	// On ordinary users a read-only directory also fails; root bypasses chmod.
	if runtime.GOOS != "windows" && os.Geteuid() != 0 {
		locked := filepath.Join(root, "locked")
		os.Mkdir(locked, 0500)
		defer os.Chmod(locked, 0700)
		if _, err := writableDirectory(filepath.Join(locked, "userdata")); err == nil {
			t.Fatal("read-only storage accepted")
		}
	}
}

func TestPortableStorageFallback(t *testing.T) {
	root, fallback := t.TempDir(), t.TempDir()
	called := false
	user := func() (string, error) { called = true; return fallback, nil }
	got, err := portableStorage(root, user)
	if err != nil || !sameDirectory(got, filepath.Join(root, "userdata")) || called {
		t.Fatal("portable first", got, err, called)
	}
	blocked := filepath.Join(t.TempDir(), "file")
	os.WriteFile(blocked, nil, 0600)
	got, err = portableStorage(blocked, user)
	if err != nil || !sameDirectory(filepath.Dir(got), fallback) || !called {
		t.Fatal("fallback", got, err, called)
	}
	if _, err := StateDirectory(filepath.Join(blocked, "explicit")); err == nil {
		t.Fatal("explicit override silently fell back")
	}
}

func TestPortableStateLeavesOriginalsReadOnly(t *testing.T) {
	root := t.TempDir()
	seed := map[string][]byte{"PINBALL.CFG": {0, 1, 0, 0, 0, 1}, "TABLE1.HI": frontend.Defaults(1).MarshalBinary()}
	for name, b := range seed {
		if err := os.WriteFile(filepath.Join(root, name), b, 0400); err != nil {
			t.Fatal(err)
		}
	}
	dir, err := portableStorage(root, func() (string, error) { t.Fatal("unexpected fallback"); return "", nil })
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckStateSeparate(root, dir); err != nil {
		t.Fatal(err)
	}
	if err := CheckStateSeparate(root, root); err == nil {
		t.Fatal("original data accepted as writable state")
	}
	config := settings.Store{Directory: dir, SeedDirectory: root}
	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := config.Save(c); err != nil {
		t.Fatal(err)
	}
	scores := frontend.FileStore{Directory: dir, SeedDirectory: root}
	s, err := scores.Load(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := scores.Save(1, s); err != nil {
		t.Fatal(err)
	}
	for name, b := range seed {
		got, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || string(got) != string(b) {
			t.Fatal("original modified", name, err)
		}
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatal("state missing from userdata", name, err)
		}
	}
}

// Windows resolves temporary directories to their canonical spelling; compare
// directory identity so case/short-name differences do not imply a storage bug.
func sameDirectory(a, b string) bool {
	left, err := os.Stat(a)
	if err != nil {
		return false
	}
	right, err := os.Stat(b)
	return err == nil && os.SameFile(left, right)
}
