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
	"testing"
)

func TestPrivateDeluxeSharedRuntime(t *testing.T) {
	d, a, b := os.Getenv("PF_DELUXE_CD_DATA"), os.Getenv("PF_RUNTIME_DATA"), os.Getenv("PF_POWERPACK_DATA")
	if d == "" || a == "" || b == "" {
		t.Skip("supply private A/B/D installations")
	}
	testDeluxeSharedRuntime(t, d, a, b, datalayout.DeluxeCDProfile)
	t.Run("Alt", func(t *testing.T) {
		c := os.Getenv("PF_DELUXE_CD_ALT_DATA")
		if c == "" {
			t.Skip("supply PF_DELUXE_CD_ALT_DATA")
		}
		testDeluxeSharedRuntime(t, c, a, b, datalayout.DeluxeCDAltProfile)
	})
}

func testDeluxeSharedRuntime(t *testing.T, d, a, b, profile string) {
	read := func(dir, name string) []byte {
		raw, e := os.ReadFile(filepath.Join(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		return raw
	}
	prepare := func(dir, name string) []byte {
		out, e := datalayout.PreparePRG(name, read(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		return out
	}
	for _, tc := range []struct {
		dir, id  string
		priority uint8
	}{{a, datalayout.RetailProfile, 1}, {b, datalayout.PowerPackProfile, 0}, {d, profile, 0}} {
		// Stage only that installation. No canonical fixture is visible to Load or
		// any of its four factories, and the optional CFG is intentionally absent.
		r := loadCompatible(t, stageInstallation(t, tc.dir))
		if r.ProfileID != tc.id {
			t.Fatal(r.ProfileID)
		}
		s, e := r.Model.Factory(tablelogic.Decimal{})
		if e != nil {
			t.Fatal(e)
		}
		g := s.(*partyland.Game)
		spec := g.Display.Jingle("S_EMPTY")
		if spec != (tablelogic.JingleSpec{Position: 62, Repeat: 0, Priority: tc.priority}) {
			t.Fatal(spec)
		}
		g.Cue("S_EMPTY")
		if g.Audio.Priority != tc.priority || g.Audio.Position != 62 || g.Audio.JumpCount != 0 {
			t.Fatal(g.Audio)
		}
		clock := tablelogic.MusicClock{Priority: 1}
		if clock.Play(spec, 62, nil) != (tc.priority == 1) {
			t.Fatal("arbitration")
		}
		if tc.id == datalayout.DeluxeCDProfile || tc.id == datalayout.DeluxeCDAltProfile {
			art := r.View.Art
			if art.Startup[0].Width != 320 || art.Startup[0].Height != 123 || art.Startup[1].Height != 117 || art.StartupLowerY != 123 {
				t.Fatal("geometry coerced")
			}
			r.Model.Segment, r.Model.SegmentTick = 1, 19
			frame := r.Frame()
			for _, p := range []struct{ x, y, asset, row int }{{17, 122, 0, 122}, {17, 123, 1, 0}, {17, 239, 1, 116}} {
				c := art.Startup[p.asset].Frame().RGBAAt(p.x, p.row)
				fade := func(v byte) byte { dac := int(v>>2) * 19 / 20; return byte(dac<<2 | dac>>4) }
				c.R, c.G, c.B = fade(c.R), fade(c.G), fade(c.B)
				if got := frame.RGBAAt(p.x, p.y); got != c {
					t.Fatal("startup placement", p, got, c)
				}
			}
		}
	}
	for _, tc := range []struct {
		name   string
		decode func([]byte) (*physics.Table, error)
	}{{"TABLE1.PRG", physics.DecodePartyLand}, {"TABLE2.PRG", physics.DecodeSpeedDevils}, {"TABLE3.PRG", physics.DecodeGameshow}, {"TABLE4.PRG", physics.DecodeStones}} {
		x, e := tc.decode(prepare(a, tc.name))
		if e != nil {
			t.Fatal(e)
		}
		y, e := tc.decode(prepare(d, tc.name))
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(x, y) {
			t.Fatal(tc.name, "graphics/physics/control differs")
		}
	}
	ca, e := assets.DecodeFrontend(prepare(a, "INTRO.PRG"))
	if e != nil {
		t.Fatal(e)
	}
	cd, e := assets.DecodeFrontend(prepare(d, "INTRO.PRG"))
	if e != nil {
		t.Fatal(e)
	}
	if ca.SidebarInfo != cd.SidebarInfo || ca.OptionsInfo != cd.OptionsInfo || !reflect.DeepEqual(ca.Tables, cd.Tables) {
		t.Fatal("sidebar/options/table cards")
	}
	for i := 2; i < 8; i++ {
		if !reflect.DeepEqual(ca.Startup[i], cd.Startup[i]) {
			t.Fatal("retained startup asset", i)
		}
	}
	for _, tc := range []struct {
		name   string
		decode func([]byte) (*audio.Module, error)
	}{{"INTRO.MOD", audio.DecodeIntro}, {"MOD2.MOD", audio.DecodeMenu}, {"TABLE1.MOD", audio.Decode}, {"TABLE2.MOD", audio.DecodeSpeedDevils}, {"TABLE3.MOD", audio.DecodeGameshow}, {"TABLE4.MOD", audio.DecodeStones}} {
		x, e := tc.decode(read(a, tc.name))
		if e != nil {
			t.Fatal(e)
		}
		y, e := tc.decode(read(d, tc.name))
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(x, y) {
			t.Fatal(tc.name, "module differs")
		}
	}
	out := prepare(d, "TABLE4.PRG")
	raw := read(d, "TABLE4.PRG")
	if bytes.Equal(out[94518:94520], raw[43475:43477]) {
		t.Fatal("raw selector leaked")
	}
}
