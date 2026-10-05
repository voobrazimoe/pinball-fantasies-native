package frontend

import (
	"bytes"
	"os"
	"path/filepath"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/datalayout"
	"reflect"
	"strings"
	"testing"
)

func TestPossessedAlternateAuditBoundary(t *testing.T) {
	alt := os.Getenv("PF_UNSUPPORTED_DATA")
	canonical := os.Getenv("PF_RUNTIME_DATA")
	if alt == "" || canonical == "" {
		t.Skip("supply both private audited installations")
	}
	read := func(dir, name string) []byte {
		b, e := os.ReadFile(filepath.Join(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	b, d := read(canonical, "TABLE1.PRG"), read(alt, "TABLE1.PRG")
	at := 0x1a9ac
	shift := 16
	if !bytes.Equal(b[at-3:at], d[at+shift-3:at+shift]) || !bytes.Equal(b[at+6:at+9], d[at+shift+6:at+shift+9]) {
		t.Fatal("S_EMPTY neighboring identity anchors changed")
	}
	if !bytes.Equal(b[at:at+2], d[at+shift:at+shift+2]) || b[at+2] == d[at+shift+2] {
		t.Fatal("expected priority-only S_EMPTY semantic difference")
	}
	// The fingerprint remains required even after the other records are relocated.
	if datalayout.Validate("TABLE1.PRG", d) == nil {
		t.Fatal("alternate accepted without adapter")
	}
	if _, e := Load(alt, nil); e == nil || !strings.Contains(e.Error(), "INTRO.PRG") {
		t.Fatalf("unsupported installation boundary: %v", e)
	}
	for name, decode := range map[string]func([]byte) (*audio.Module, error){"INTRO.MOD": audio.DecodeIntro, "MOD2.MOD": audio.DecodeMenu, "TABLE1.MOD": audio.Decode, "TABLE2.MOD": audio.DecodeSpeedDevils, "TABLE3.MOD": audio.DecodeGameshow, "TABLE4.MOD": audio.DecodeStones} {
		c, e := decode(read(canonical, name))
		if e != nil {
			t.Fatal(e)
		}
		a, e := decode(read(alt, name))
		if e != nil {
			t.Fatal(name, e)
		}
		if !reflect.DeepEqual(c, a) {
			t.Fatal(name, "decoded module differs")
		}
	}
}

func TestMalformedModulesRetainFilenameContext(t *testing.T) {
	source := originalInstallation(t)
	for _, name := range []string{"INTRO.MOD", "MOD2.MOD", "TABLE1.MOD", "TABLE2.MOD", "TABLE3.MOD", "TABLE4.MOD"} {
		t.Run(name, func(t *testing.T) {
			dir := stageInstallation(t, source)
			if e := os.WriteFile(filepath.Join(dir, name), []byte{0}, 0600); e != nil {
				t.Fatal(e)
			}
			if _, e := Load(dir, nil); e == nil || !strings.Contains(e.Error(), name+":") {
				t.Fatalf("missing failing module context: %v", e)
			}
		})
	}
}
