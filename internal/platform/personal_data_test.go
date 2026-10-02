package platform

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func personalTestZip(t *testing.T, names ...string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for _, name := range names {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = f.Write([]byte("original seed")); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestPersonalExtractionReadonlyAndCleanup(t *testing.T) {
	dir, cleanup, err := extractPersonalData(personalTestZip(t, "PINBALL.CFG", "TABLE1.PRG"))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, name := range []string{"PINBALL.CFG", "TABLE1.PRG"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(b) != "original seed" {
			t.Fatalf("seed %s: %q %v", name, b, err)
		}
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || info.Mode().Perm()&0200 != 0 {
			t.Fatalf("seed is writable: %v %v", info, err)
		}
	}
	cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("temporary originals not removed: %v", err)
	}
}

func TestPersonalExtractionRejectsInvalidEntries(t *testing.T) {
	for _, names := range [][]string{{"../outside"}, {"nested/file"}, {"C:escape"}, {"same", "same"}, {""}} {
		t.Run(names[0], func(t *testing.T) {
			if dir, cleanup, err := extractPersonalData(personalTestZip(t, names...)); err == nil {
				cleanup()
				t.Fatalf("accepted invalid archive at %s", dir)
			}
		})
	}
	if _, _, err := extractPersonalData([]byte("invalid zip")); err == nil {
		t.Fatal("accepted malformed archive")
	}
}

func TestPersonalExplicitDataOverride(t *testing.T) {
	old := personalPayload
	personalPayload = personalTestZip(t, "PINBALL.CFG")
	defer func() { personalPayload = old }()
	dir := t.TempDir()
	got, cleanup, err := PrepareDataDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if got != dir {
		t.Fatalf("explicit override lost: %s", got)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatal("bundle extracted into override")
	}
}

func TestPersonalRejectedArchiveCleanupStaysInsideTemp(t *testing.T) {
	parent := t.TempDir()
	t.Setenv("TMPDIR", parent)
	t.Setenv("TMP", parent)
	outside := filepath.Join(parent, "outside")
	if err := os.WriteFile(outside, []byte("untouched"), 0400); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(outside, 0600)
	if _, _, err := extractPersonalData(personalTestZip(t, "valid", "../outside")); err == nil {
		t.Fatal("accepted traversal")
	}
	info, err := os.Stat(outside)
	if err != nil || info.Mode().Perm()&0200 != 0 {
		t.Fatalf("cleanup touched outside file: %v %v", info, err)
	}
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 1 || entries[0].Name() != "outside" {
		t.Fatalf("partial extraction leaked: %v %v", entries, err)
	}
}
