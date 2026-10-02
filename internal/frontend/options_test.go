package frontend

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func configured(t *testing.T, c settings.Config) *Runtime {
	t.Helper()
	s := &settings.Store{Directory: t.TempDir()}
	if e := s.Save(c); e != nil {
		t.Fatal(e)
	}
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../PINBALL.CFG", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD")
	r, e := LoadConfigured("../..", nil, s)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func runKey(t *testing.T, r *Runtime, k Key) {
	t.Helper()
	if e := r.Update(Input{Keys: []Key{k}}); e != nil {
		t.Fatal(e)
	}
}
func runTicks(t *testing.T, r *Runtime, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if e := r.Update(Input{}); e != nil {
			t.Fatal(e)
		}
	}
}
func checkpoint(t *testing.T, name string, im *image.RGBA) {
	t.Helper()
	if strings.HasPrefix(name, "off-table") {
		raw, e := os.ReadFile("../../analysis/pf11.2-full-table-fixtures.json")
		if e != nil {
			t.Fatal(e)
		}
		var fixtures map[string]string
		if e = json.Unmarshal(raw, &fixtures); e != nil {
			t.Fatal(e)
		}
		got := fmt.Sprintf("%x", sha256.Sum256(im.Pix))
		if fixtures[name] != got {
			t.Fatalf("checkpoint %s hash %s want %s", name, got, fixtures[name])
		}
	}
	dir := os.Getenv("PF11_CHECKPOINT_DIR")
	if dir == "" {
		return
	}
	if e := os.MkdirAll(dir, 0755); e != nil {
		t.Fatal(e)
	}
	f, e := os.Create(filepath.Join(dir, name+".png"))
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if e = png.Encode(f, im); e != nil {
		t.Fatal(e)
	}
}
func TestOptionsControlsPersistenceAndAudioContinuity(t *testing.T) {
	for _, menu := range []bool{false, true} {
		r := configured(t, settings.Defaults())
		runKey(t, r, Space)
		if menu {
			r.Player = audio.New(r.Menu)
		}
		source := r.AudioSource()
		control := *r.Player
		for cycle := 0; cycle < 3; cycle++ {
			input := []Key{F5}
			for i := 0; i < 46; i++ {
				input = append(input, 0)
			}
			input = append(input, Up, Down)
			for i := 0; i < 5; i++ {
				input = append(input, Enter, Down)
			}
			input = append(input, Enter)
			for i := 0; i < 41; i++ {
				input = append(input, 0)
			}
			for _, k := range input {
				in := Input{}
				if k != 0 {
					in.Keys = []Key{k}
				}
				if e := r.Update(in); e != nil {
					t.Fatal(e)
				}
				expected := control.Render(audio.Rate / 60)
				if r.AudioSource() != source || !bytes.Equal(expected, r.PCM) || !reflect.DeepEqual(control, *r.Player) {
					t.Fatal("visual options transition changed producer/phase")
				}
				if r.Model.Mode == Options {
					checkpoint(t, fmt.Sprintf("options-row%d", r.Model.OptionRow), r.Frame())
				}
			}
			if r.Model.Mode != Selector || r.Model.Selected != 1 || r.Model.Session != nil {
				t.Fatal("return lifecycle", r.Model.Mode)
			}
		}
		// Esc keeps values; SAVE AND EXIT doesn't itself write to disk.
		before, _ := r.Model.SettingsStore.Load()
		if before != settings.Defaults() {
			t.Fatal("options prematurely persisted")
		}
		runKey(t, r, F5)
		runTicks(t, r, 46)
		checkpoint(t, "options-changed", r.Frame())
		runKey(t, r, Escape)
		runTicks(t, r, 40)
		checkpoint(t, "return", r.Frame())
		runKey(t, r, F1)
		saved, e := r.Model.SettingsStore.Load()
		if e != nil || saved != r.Model.Settings {
			t.Fatal(saved, e)
		}
		r2, e := LoadConfigured("../..", nil, r.Model.SettingsStore)
		if e != nil || r2.Model.Settings != saved {
			t.Fatal("restart", e)
		}
	}
}
func gameParts(s Session) (*physics.Game, *presentation.Display, bool) {
	switch g := s.(type) {
	case *partyland.Game:
		return g.Physics, g.Display, g.MusicOff
	case *speeddevils.Game:
		return g.Physics, g.Display, g.MusicOff
	case *gameshow.Game:
		return g.Physics, g.Display, g.MusicOff
	case *stones.Game:
		return g.Physics, g.Display, g.MusicOff
	}
	panic("table type")
}
func TestFourTableSettings(t *testing.T) {
	variants := []settings.Config{settings.Defaults(), settings.Legacy(), {Balls: 1, Angle: 1, ScrollMode: 0, Music: 1, Resolution: 0}, {Balls: 1, ScrollMode: 1, Resolution: 1}}
	for table := 1; table <= 4; table++ {
		for n, c := range variants {
			t.Run(fmt.Sprintf("table%d-config%d", table, n), func(t *testing.T) {
				r := configured(t, c)
				runKey(t, r, Space)
				runKey(t, r, Key(int(F1)+table-1))
				checkpoint(t, fmt.Sprintf("table%d-mode%d-attract", table, c.Resolution), r.Frame())
				runTicks(t, r, 3)
				runKey(t, r, Enter)
				session := r.Model.Session
				g, d, off := gameParts(session)
				typed := session.(interface {
					SessionSettings() settings.Config
					BaseBalls() byte
				})
				balls := byte(3)
				if c.Balls == 1 {
					balls = 5
				}
				if typed.SessionSettings() != c || typed.BaseBalls() != balls || off != (c.Music == 1) {
					t.Fatal("session propagation", typed.SessionSettings(), typed.BaseBalls(), off)
				}
				if got := r.Frame().Rect.Size(); got != image.Pt(320, c.FieldHeight()+33) {
					t.Fatal("logical geometry", got)
				}
				checkpoint(t, fmt.Sprintf("table%d-config%d-chute", table, n), r.Frame())
				if dir := os.Getenv("PF11_CHECKPOINT_DIR"); dir != "" {
					var palette [768]byte
					switch native := session.(type) {
					case *partyland.Game:
						palette = native.Palette()
					case *speeddevils.Game:
						palette = native.Palette()
					case *gameshow.Game:
						palette = native.Palette()
					case *stones.Game:
						palette = native.Palette()
					}
					data, _ := json.Marshal(palette)
					if e := os.WriteFile(filepath.Join(dir, fmt.Sprintf("table%d-config%d-palette.json", table, n)), data, 0600); e != nil {
						t.Fatal(e)
					}
				}
				g.SpringPosition = 32
				checkpoint(t, fmt.Sprintf("table%d-mode%d-spring", table, c.Resolution), r.Frame())
				d.Text("OPTIONS FIXTURE", 0, 0, 13)
				checkpoint(t, fmt.Sprintf("table%d-mode%d-text", table, c.Resolution), settingsMatrixFrame(session, g, d, table))
				// Original bitmap operation, with table-owned content and unchanged scheduler.
				labels := make([]string, 0, len(d.Content.Animations))
				for label := range d.Content.Animations {
					labels = append(labels, label)
				}
				sort.Strings(labels)
				for _, label := range labels {
					anim := d.Content.Animations[label]
					if len(anim.Offsets) > 0 {
						d.Bitmap(anim.Offsets[0])
						break
					}
				}
				checkpoint(t, fmt.Sprintf("table%d-mode%d-bitmap", table, c.Resolution), settingsMatrixFrame(session, g, d, table))
				for i := 0; i < 8; i++ {
					if e := g.Sync(physics.Inputs{}); e != nil {
						t.Fatal(e)
					}
				}
				if g.Raster != (g.BottomRaster()+33)*16 {
					t.Fatal("initial forced/chute raster", g.Raster)
				}
				// Gameplay M remains temporary; table reload restores the config.
				runKey(t, r, 50)
				_, _, nowOff := gameParts(session)
				if nowOff == off || r.Model.Settings != c {
					t.Fatal("runtime M persisted")
				}
				r.Model.Settings.Cycle(0)
				if typed.BaseBalls() != balls {
					t.Fatal("mutated live ball limit")
				}
				r.Model.Settings = c
				// Another game in the loaded table retains runtime M.
				r.Model.Mode = TableAttract
				runKey(t, r, Enter)
				_, _, retained := gameParts(r.Model.Session)
				if retained != nowOff {
					t.Fatal("M lost on game restart")
				}
				runKey(t, r, P)
				runKey(t, r, Escape)
				runKey(t, r, Key(21))
				if r.Model.Mode != Selector {
					t.Fatal("abort return", r.Model.Mode)
				}
				runKey(t, r, Key(int(F1)+table-1))
				runKey(t, r, Enter)
				_, _, off = gameParts(r.Model.Session)
				if off != (c.Music == 1) {
					t.Fatal("fresh session M state")
				}
				checkpoint(t, fmt.Sprintf("table%d-mode%d-gameplay", table, c.Resolution), r.Frame())
			})
		}
	}
}

func settingsMatrixFrame(session Session, g *physics.Game, d *presentation.Display, table int) *image.RGBA {
	var p [768]byte
	switch native := session.(type) {
	case *partyland.Game:
		p = native.Palette()
	case *speeddevils.Game:
		p = native.Palette()
	case *gameshow.Game:
		p = native.Palette()
	case *stones.Game:
		p = native.Palette()
	}
	on := [4]byte{242, 128, 153, 79}[table-1]
	off := [4]byte{96, 98, 114, 231}[table-1]
	if table != 4 {
		p = presentation.MatrixPaletteMode(p, g.ReferenceMode, on)
	}
	return presentation.ComposeNative(g.FramePalette(p), d, p, off, on, g.Settings, g.ScreenOffset)
}
func TestFourTableBothModesGameplayAndCapture(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for mode := byte(0); mode < 2; mode++ {
			c := settings.Legacy()
			c.Resolution = mode
			if table == 4 {
				c.Angle = 0
			}
			r := configured(t, c)
			runKey(t, r, Space)
			runKey(t, r, Key(int(F1)+table-1))
			runKey(t, r, Enter)
			runTicks(t, r, 220)
			r.Model.Session.Release(12, 7)
			runTicks(t, r, 120)
			checkpoint(t, fmt.Sprintf("table%d-mode%d-gameplay", table, mode), r.Frame())
			g, _, _ := gameParts(r.Model.Session)
			ball := g.Ball
			clock := g.Syncs
			g.ScrollForce = func() int16 { return 187 } // source SCREENFORCE remains authoritative
			// Apply just the camera by a held sync; table task/collision clock belongs to Sync.
			g.Ball.Hold = true
			if e := g.Sync(physics.Inputs{}); e != nil {
				t.Fatal(e)
			}
			if g.Raster != (187+33)*16 {
				t.Fatal("forced capture", table, mode, g.Raster)
			}
			checkpoint(t, fmt.Sprintf("table%d-mode%d-force", table, mode), r.Frame())
			if g.Settings != c || g.Syncs != clock+1 {
				t.Fatal("presentation changed schedule")
			}
			g.Ball = ball
		}
	}
}
func TestOptionsFromSelectorText(t *testing.T) {
	r := configured(t, settings.Defaults())
	runKey(t, r, Space)
	runTicks(t, r, 50)
	runKey(t, r, Space)
	runTicks(t, r, 45)
	producer := r.AudioSource()
	runKey(t, r, F5)
	if r.Model.Mode != SelectorText || !r.Model.OptionsPending {
		t.Fatal("text-page F5 did not queue menu")
	}
	runTicks(t, r, 21)
	if r.Model.Mode != Options || r.AudioSource() != producer {
		t.Fatal("queued menu lifecycle")
	}
}

func TestOptionsVisualRowsAndDefaults(t *testing.T) {
	r := configured(t, settings.Defaults())
	runKey(t, r, Space)
	runTicks(t, r, 50)
	runKey(t, r, F5)
	runTicks(t, r, 46)
	checkpoint(t, "options-defaults", r.Frame())
	for row := 0; row < 6; row++ {
		if r.Model.OptionRow != row {
			t.Fatal("cursor order", row)
		}
		checkpoint(t, fmt.Sprintf("options-selected%d", row), r.Frame())
		if row < 5 {
			runKey(t, r, Down)
		}
	}
	runKey(t, r, Up)
	runKey(t, r, Space)
	if r.Model.Settings.Resolution != 1 {
		t.Fatal("Space did not toggle")
	}
}

func TestOptionsPersistOnProgramClose(t *testing.T) {
	r := configured(t, settings.Defaults())
	runKey(t, r, Space)
	runKey(t, r, F5)
	runTicks(t, r, 46)
	runKey(t, r, Space)
	runKey(t, r, Escape)
	runTicks(t, r, 40)
	if e := r.Update(Input{Close: true}); e != nil {
		t.Fatal(e)
	}
	c, e := r.Model.SettingsStore.Load()
	if e != nil || c.Balls != 1 || r.Model.Mode != Quit {
		t.Fatal(c, e)
	}
}
