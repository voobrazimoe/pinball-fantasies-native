package dosname

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestReadFileIgnoresCase(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "intro.prg", "lower")
	b, err := ReadFile(dir, "INTRO.PRG")
	if err != nil || string(b) != "lower" {
		t.Fatalf("%q %v", b, err)
	}
	if _, err := Stat(dir, "INTRO.PRG"); err != nil {
		t.Fatal(err)
	}
}

func TestMissingKeepsNotExist(t *testing.T) {
	_, err := ReadFile(t.TempDir(), "TABLE1.PRG")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	if _, err := ReadFile(filepath.Join(t.TempDir(), "missing"), "TABLE1.PRG"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
}

func TestAmbiguousCaseIsRejected(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "Table1.prg", "a")
	write(t, dir, "table1.PRG", "b")
	entries, _ := os.ReadDir(dir)
	if len(entries) < 2 {
		t.Skip("case-insensitive file system")
	}
	if _, err := ReadFile(dir, "TABLE1.PRG"); err == nil {
		t.Fatal("ambiguous names accepted")
	}
}

func TestExactNameWins(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "MOD2.MOD", "exact")
	write(t, dir, "mod2.mod", "lower")
	if entries, _ := os.ReadDir(dir); len(entries) < 2 {
		t.Skip("case-insensitive file system")
	}
	b, err := ReadFile(dir, "MOD2.MOD")
	if err != nil || string(b) != "exact" {
		t.Fatalf("%q %v", b, err)
	}
}
