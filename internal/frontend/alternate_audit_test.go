package frontend

import (
	"bytes"
	"os"
	"path/filepath"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/tablelogic"
	"reflect"
	"strings"
	"testing"
)

func TestPossessedPowerPackRuntime(t *testing.T) {
	alt := os.Getenv("PF_POWERPACK_DATA")
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
	for _, tc := range []struct {
		dir, profile string
		priority     uint8
	}{
		{canonical, datalayout.RetailProfile, 1}, {alt, datalayout.PowerPackProfile, 0},
	} {
		r := loadCompatible(t, stageInstallation(t, tc.dir))
		if r.ProfileID != tc.profile {
			t.Fatal(r.ProfileID)
		}
		session, err := r.Model.Factory(tablelogic.Decimal{})
		if err != nil {
			t.Fatal(err)
		}
		g := session.(*partyland.Game)
		spec := g.Display.Jingle("S_EMPTY")
		if spec != (tablelogic.JingleSpec{Position: 62, Repeat: 0, Priority: tc.priority}) {
			t.Fatal(spec)
		}
		g.Cue("S_EMPTY")
		if g.Audio.Priority != tc.priority || g.Audio.Position != 62 || g.Audio.JumpCount != 0 {
			t.Fatal("cue semantics lost", g.Audio)
		}
		clock := tablelogic.MusicClock{Priority: 1}
		if clock.Play(spec, 62, nil) != (tc.priority == 1) {
			t.Fatal("shared arbitration lost edition priority")
		}
	}
	// Shared graphic and physics decoders must produce identical models. The
	// only proved edition semantic difference lives in the matrix jingle record.
	for _, name := range []string{"INTRO.PRG", "TABLE1.PRG", "TABLE2.PRG", "TABLE3.PRG", "TABLE4.PRG"} {
		ca, err := datalayout.PreparePRG(name, read(canonical, name))
		if err != nil {
			t.Fatal(err)
		}
		pb, err := datalayout.PreparePRG(name, read(alt, name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "INTRO.PRG" {
			x, err := assets.DecodeFrontend(ca)
			if err != nil {
				t.Fatal(err)
			}
			y, err := assets.DecodeFrontend(pb)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(x, y) {
				t.Fatal("INTRO assets/geometry/sidebar differ")
			}
		} else {
			decode := map[string]func([]byte) (*physics.Table, error){"TABLE1.PRG": physics.DecodePartyLand, "TABLE2.PRG": physics.DecodeSpeedDevils, "TABLE3.PRG": physics.DecodeGameshow, "TABLE4.PRG": physics.DecodeStones}[name]
			x, err := decode(ca)
			if err != nil {
				t.Fatal(err)
			}
			y, err := decode(pb)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(x, y) {
				t.Fatal(name, "graphics/physics/control model differs")
			}
		}
	}
	for mask := 1; mask < 7; mask++ {
		dir := stageInstallation(t, canonical)
		for i, name := range []string{"INTRO.PRG", "TABLE1.PRG", "TABLE2.PRG"} {
			if mask&(1<<i) != 0 {
				if err := os.WriteFile(filepath.Join(dir, name), read(alt, name), 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
		if _, err := Load(dir, nil); err == nil || !strings.Contains(err.Error(), "coherent") {
			t.Fatal("frontend hybrid accepted", mask, err)
		}
	}
	for _, tc := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"unknown", func(b []byte) []byte { return append([]byte{0}, b...) }},
		{"malformed", func(b []byte) []byte { return b[:100] }},
		{"incomplete", func(b []byte) []byte { return nil }},
	} {
		dir := stageInstallation(t, alt)
		path := filepath.Join(dir, "INTRO.PRG")
		if tc.name == "incomplete" {
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
		} else if err := os.WriteFile(path, tc.mutate(read(alt, "INTRO.PRG")), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(dir, nil); err == nil {
			t.Fatal(tc.name, "accepted")
		}
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
