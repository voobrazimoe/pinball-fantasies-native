package settings

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRecordAndPersistence(t *testing.T) {
	dir, seed := t.TempDir(), t.TempDir()
	s := Store{dir, seed}
	c, e := s.Load()
	if e != nil || c != Defaults() {
		t.Fatal(c, e)
	}
	source := []byte{1, 1, 2, 1, 1, 0}
	os.WriteFile(filepath.Join(seed, "PINBALL.CFG"), source, 0600)
	c, e = s.Load()
	if e != nil || c != (Config{1, 1, ScrollSoft, 1, 1}) {
		t.Fatal(c, e)
	}
	c.Cycle(2)
	if e = s.Save(c); e != nil {
		t.Fatal(e)
	}
	got, e := s.Load()
	if e != nil || got != c {
		t.Fatal(got, e)
	}
	b, _ := os.ReadFile(filepath.Join(seed, "PINBALL.CFG"))
	if !bytes.Equal(b, source) {
		t.Fatal("modified seed")
	}
	for _, bad := range [][]byte{nil, {0, 0, 1}, {2, 0, 1, 0, 0, 0}, {0, 0, 3, 0, 0, 0}, {0, 0, 1, 0, 0, 2}, make([]byte, 7)} {
		os.WriteFile(filepath.Join(dir, "PINBALL.CFG"), bad, 0600)
		got, e = s.Load()
		if e != nil || got != Defaults() {
			t.Fatal("fallback", got, e)
		}
	}
	for row := 0; row < 5; row++ {
		c = Defaults()
		initial := c
		n := 2
		if row == 2 {
			n = 4
		}
		for i := 0; i < n; i++ {
			c.Cycle(row)
		}
		if c != initial {
			t.Fatal("cycle", row, c)
		}
	}
}
func TestMonoDACIntegerMean(t *testing.T) {
	var p [768]byte
	p[0], p[1], p[2] = 252, 128, 4
	p = MonoDAC(p)
	if p[0] != 130 || p[1] != 130 || p[2] != 130 {
		t.Fatal(p[:3])
	}
}

func TestNativeVersionMigrationValidation(t *testing.T) {
	for mode := byte(0); mode < 2; mode++ {
		raw := []byte{1, 1, 2, 1, 1, mode}
		legacy, e := DecodeLegacy(raw)
		if e != nil || legacy.Mode != mode {
			t.Fatal(legacy, e)
		}
		c, e := Decode(raw)
		if e != nil || c != (Config{1, 1, ScrollSoft, 1, 1}) {
			t.Fatal(c, e)
		}
		if string(c.Bytes()[:4]) != "PFNC" {
			t.Fatal("legacy written")
		}
	}
	for scroll := ScrollHard; scroll <= ScrollOff; scroll++ {
		c := Legacy()
		c.ScrollMode = scroll
		got, e := Decode(c.Bytes())
		if e != nil || got != c {
			t.Fatal(got, e)
		}
		for n := 0; n < len(c.Bytes()); n++ {
			if _, e := Decode(c.Bytes()[:n]); e == nil {
				t.Fatal("truncation accepted", n)
			}
		}
	}
	for _, raw := range [][]byte{{'P', 'F', 'N', 'C', 2, 5, 0, 0, 0, 0, 0}, {'P', 'F', 'N', 'C', 1, 5, 0, 0, 4, 0, 0}, {'P', 'F', 'N', 'C', 1, 4, 0, 0, 0, 0, 0}, {'P', 'F', 'N', 'C', 1, 5, 0, 0, 0, 0, 0, 0}} {
		if _, e := Decode(raw); e == nil {
			t.Fatal("bad native accepted", raw)
		}
	}
}

func TestNativeStoreFallbackAndInvalidSave(t *testing.T) {
	s := Store{Directory: t.TempDir(), SeedDirectory: t.TempDir()}
	future := []byte{'P', 'F', 'N', 'C', 9, 5, 0, 0, 0, 0, 0}
	path := filepath.Join(s.Directory, "PINBALL.CFG")
	if e := os.WriteFile(path, future, 0600); e != nil {
		t.Fatal(e)
	}
	c, e := s.Load()
	if e != nil || c != Defaults() {
		t.Fatal("future fallback", c, e)
	}
	unchanged, _ := os.ReadFile(path)
	if !bytes.Equal(unchanged, future) {
		t.Fatal("load rewrote future format")
	}
	if e = s.Save(Config{ScrollMode: 4}); e == nil {
		t.Fatal("invalid save accepted")
	}
	unchanged, _ = os.ReadFile(path)
	if !bytes.Equal(unchanged, future) {
		t.Fatal("invalid save replaced config")
	}
	native := Config{Balls: 1, Angle: 1, ScrollMode: ScrollOff, Music: 1, Resolution: 1}
	if e = s.Save(native); e != nil {
		t.Fatal(e)
	}
	raw, _ := os.ReadFile(path)
	if !bytes.Equal(raw, []byte{'P', 'F', 'N', 'C', 1, 5, 1, 1, 3, 1, 1}) {
		t.Fatal("native wire encoding", raw)
	}
	got, e := s.Load()
	if e != nil || got != native {
		t.Fatal(got, e)
	}
	entries, _ := os.ReadDir(s.Directory)
	if len(entries) != 1 {
		t.Fatal("atomic temporary file leak")
	}
}
