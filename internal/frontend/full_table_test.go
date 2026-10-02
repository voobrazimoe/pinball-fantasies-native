package frontend

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/settings"
	"reflect"
	"sort"
	"testing"
)

func offSession(t *testing.T, table int) (*Runtime, *physics.Game, *presentation.Display) {
	c := settings.Legacy()
	c.ScrollMode = settings.ScrollOff
	r := configured(t, c)
	runKey(t, r, Space)
	runKey(t, r, Key(int(F1)+table-1))
	checkpoint(t, fmt.Sprintf("off-table%d-attract", table), r.Frame())
	runKey(t, r, Enter)
	g, d, _ := gameParts(r.Model.Session)
	return r, g, d
}

func TestFullTableFourTableCheckpoints(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			r, g, d := offSession(t, table)
			if len(g.Table.Initial.Playfield.Indices) != 320*576 {
				t.Fatal("table extent")
			}
			check := func(name string) {
				frame := r.Frame()
				if frame.Rect.Size() != image.Pt(320, 609) {
					t.Fatal(frame.Rect)
				}
				// Compare every row against the existing viewport renderer. Rendering stays
				// in table space: only the final composition adds 33 to the output Y.
				p := paletteOf(r.Model.Session)
				if table != 4 {
					p = presentation.MatrixPaletteMode(p, g.ReferenceMode, [4]byte{242, 128, 153, 79}[table-1])
				}
				full := g.FramePalette(p)
				cfg, raster, offset := g.Settings, g.Raster, g.ScreenOffset
				g.Settings.ScrollMode = settings.ScrollSoft
				g.ScreenOffset = 0
				for y := 0; y < 576; y++ {
					g.Raster = int16((y + 33) * 16)
					view := g.FramePalette(p)
					if !bytes.Equal(full.Pix[y*full.Stride:(y+1)*full.Stride], view.Pix[:view.Stride]) {
						t.Fatalf("%s table row %d", name, y)
					}
					if !bytes.Equal(frame.Pix[(y+33)*frame.Stride:(y+34)*frame.Stride], full.Pix[y*full.Stride:(y+1)*full.Stride]) {
						t.Fatalf("%s output row %d", name, y+33)
					}
				}
				g.Settings, g.Raster, g.ScreenOffset = cfg, raster, offset
				checkpoint(t, fmt.Sprintf("off-table%d-%s", table, name), frame)
				t.Logf("%s %x", name, sha256.Sum256(frame.Pix))
			}
			check("initial-chute")
			runTicks(t, r, 220)
			r.Model.Session.Release(32, 0)
			runTicks(t, r, 120)
			check("gameplay")
			for _, pos := range []struct {
				name string
				y    int16
			}{{"upper", 80}, {"lower", 520}} {
				g.SetBall(150, pos.y, 0, 0, false)
				check(pos.name)
			}
			g.SpringPosition = 32
			check("spring")
			d.Clear()
			d.Text("FULL TABLE", 0, 0, 13)
			matrixCheck := func(name string) {
				frame := settingsMatrixFrame(r.Model.Session, g, d, table)
				p := paletteOf(r.Model.Session)
				on := [4]byte{242, 128, 153, 79}[table-1]
				off := [4]byte{96, 98, 114, 231}[table-1]
				if table != 4 {
					p = presentation.MatrixPaletteMode(p, 0, on)
				}
				bottom := image.NewRGBA(image.Rect(0, 0, 320, 240))
				d.Draw(bottom, p, off, on)
				if !bytes.Equal(frame.Pix[:33*frame.Stride], bottom.Pix[207*bottom.Stride:]) {
					t.Fatal("matrix rows changed", name)
				}
				checkpoint(t, fmt.Sprintf("off-table%d-%s", table, name), frame)
			}
			matrixCheck("matrix-text")
			labels := []string{}
			for label := range d.Content.Animations {
				labels = append(labels, label)
			}
			sort.Strings(labels)
			for _, label := range labels {
				a := d.Content.Animations[label]
				if len(a.Offsets) > 0 {
					d.Bitmap(a.Offsets[0])
					break
				}
			}
			matrixCheck("matrix-bitmap")
			g.Ball.Hold = true
			check("capture")

			g.TargetRaster = 17
			g.ScrollForce = func() int16 { return 187 }
			if e := g.Sync(physics.Inputs{}); e != nil {
				t.Fatal(e)
			}
			if g.Raster != (187+33)*16 {
				t.Fatal("source SCREENFORCE stopped updating")
			}
			before := g.FramePalette(paletteOf(r.Model.Session))
			g.Raster = 528
			if !bytes.Equal(before.Pix, g.FramePalette(paletteOf(r.Model.Session)).Pix) {
				t.Fatal("forced camera moved full table")
			}
			g.ScrollForce = nil
			g.TargetRaster = 33
			if e := g.Sync(physics.Inputs{}); e != nil {
				t.Fatal(e)
			}
			before = g.FramePalette(paletteOf(r.Model.Session))
			g.Raster = 9000
			if !bytes.Equal(before.Pix, g.FramePalette(paletteOf(r.Model.Session)).Pix) {
				t.Fatal("SCREENFORCE2 moved full table")
			}
			check("forced-camera")
		})
	}
}

func paletteOf(s Session) [768]byte { return s.(interface{ Palette() [768]byte }).Palette() }

// Compare private rule, scheduler, capture, matrix and music state too. Function
// closures have different owners; compare their installed/empty slots instead.
func equalState(t *testing.T, a, b reflect.Value, path string) {
	t.Helper()
	if path == "session.Physics.Settings.ScrollMode" {
		return
	}
	if a.Type() != b.Type() {
		t.Fatal(path, "type")
	}
	switch a.Kind() {
	case reflect.Pointer, reflect.Interface:
		if a.IsNil() != b.IsNil() {
			t.Fatal(path, "nil")
		}
		if !a.IsNil() {
			equalState(t, a.Elem(), b.Elem(), path)
		}
	case reflect.Func:
		if a.IsNil() != b.IsNil() || a.Pointer() != b.Pointer() {
			t.Fatal(path, "task/callback slot")
		}
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			equalState(t, a.Field(i), b.Field(i), path+"."+a.Type().Field(i).Name)
		}
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			t.Fatal(path, "length")
		}
		for i := 0; i < a.Len(); i++ {
			equalState(t, a.Index(i), b.Index(i), fmt.Sprintf("%s[%d]", path, i))
		}
	case reflect.Map:
		if a.Len() != b.Len() {
			t.Fatal(path, "map length")
		}
		for _, k := range a.MapKeys() {
			v := b.MapIndex(k)
			if !v.IsValid() {
				t.Fatal(path, "key")
			}
			equalState(t, a.MapIndex(k), v, path+"[map]")
		}
	case reflect.Bool:
		if a.Bool() != b.Bool() {
			t.Fatal(path, a.Bool(), b.Bool())
		}
	case reflect.String:
		if a.String() != b.String() {
			t.Fatal(path, a.String(), b.String())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a.Int() != b.Int() {
			t.Fatal(path, a.Int(), b.Int())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if a.Uint() != b.Uint() {
			t.Fatal(path, a.Uint(), b.Uint())
		}
	default:
		t.Fatal("unhandled state", path, a.Kind())
	}
}

func TestSoftOffGameplayEquivalenceAllTables(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			a, _, _ := offSession(t, table)
			b := configured(t, settings.Legacy())
			runKey(t, b, Space)
			runKey(t, b, Key(int(F1)+table-1))
			runKey(t, b, Enter)
			for tick := 0; tick < 1200; tick++ {
				if tick == 100 || tick == 700 {
					a.Model.Session.Release(32, 0)
					b.Model.Session.Release(32, 0)
				}
				input := physics.Inputs{Left: tick%93 < 18, Right: tick%71 < 14}
				if e := a.Model.Session.Sync(input); e != nil {
					t.Fatal(e)
				}
				if e := b.Model.Session.Sync(input); e != nil {
					t.Fatal(e)
				}
				if tick%71 == 0 {
					a.Model.Session.Frame()
					b.Model.Session.Frame()
				}
			}
			equalState(t, reflect.ValueOf(a.Model.Session), reflect.ValueOf(b.Model.Session), "session")
			score, _ := a.Model.Session.Result()
			t.Logf("1200 ticks score=%s", score)
		})
	}
}

func TestOffOptionsResolutionPersistence(t *testing.T) {
	r := configured(t, settings.Config{ScrollMode: settings.ScrollHard})
	runKey(t, r, Space)
	runKey(t, r, F5)
	runTicks(t, r, 46)
	runKey(t, r, Down)
	runKey(t, r, Down)
	for _, want := range []settings.ScrollMode{settings.ScrollMedium, settings.ScrollSoft, settings.ScrollOff} {
		runKey(t, r, Enter)
		if r.Model.Settings.ScrollMode != want {
			t.Fatal("scroll cycle")
		}
	}
	checkpoint(t, "options-off", r.Frame())
	runKey(t, r, Down)
	runKey(t, r, Down)
	runKey(t, r, Enter)
	if r.Model.Settings.Resolution != 1 {
		t.Fatal("HIGH")
	}
	runKey(t, r, Enter)
	if r.Model.Settings.Resolution != 0 {
		t.Fatal("NORMAL")
	}
	runKey(t, r, Enter)
	runKey(t, r, Down)
	runKey(t, r, Enter)
	runTicks(t, r, 40)
	runKey(t, r, F1)
	if r.Frame().Rect.Dy() != 609 {
		t.Fatal("OFF")
	}
	runKey(t, r, Escape)
	runKey(t, r, Key(21))
	runKey(t, r, F5)
	runTicks(t, r, 46)
	if r.Model.Settings.Values()[2] != "OFF   " {
		t.Fatal("OFF lost")
	}
	c, e := r.Model.SettingsStore.Load()
	if e != nil || c.ScrollMode != settings.ScrollOff || c.Resolution != 1 {
		t.Fatal(c, e)
	}
	runKey(t, r, Down)
	runKey(t, r, Down)
	for i := 0; i < 3; i++ {
		runKey(t, r, Enter)
	} // OFF -> HARD -> MEDIUM -> SOFT
	runKey(t, r, Escape)
	runTicks(t, r, 40)
	runKey(t, r, F1)
	if r.Frame().Rect.Dy() != 350 {
		t.Fatal("HIGH not restored")
	}
}

// Exercise upper-plane masking and animated flipper deltas through the shared
// table-space compositor, then compare the full frame with viewport samples.
func TestFullTableDynamicLayers(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			r, g, _ := offSession(t, table)
			runTicks(t, r, 220)
			g.SetBall(150, 80, 0, 0, true)
			g.Ball.Hold = true
			for n := 0; n < 8; n++ {
				if e := g.Sync(physics.Inputs{Left: true, Right: true}); e != nil {
					t.Fatal(e)
				}
			}
			p := paletteOf(r.Model.Session)
			full := g.FramePalette(p)
			cfg, raster := g.Settings, g.Raster
			g.Settings.ScrollMode = settings.ScrollSoft
			for y := 0; y < 576; y++ {
				g.Raster = int16((y + 33) * 16)
				viewport := g.FramePalette(p)
				if !bytes.Equal(full.Pix[y*full.Stride:(y+1)*full.Stride], viewport.Pix[:viewport.Stride]) {
					t.Fatal("upper-plane/flipper row", y)
				}
			}
			g.Settings, g.Raster = cfg, raster
			checkpoint(t, fmt.Sprintf("full-table%d-upper-plane-flippers", table), r.Frame())
		})
	}
}

func TestOffNudgePhysicsAndTiltAllTables(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for _, scroll := range []settings.ScrollMode{settings.ScrollHard, settings.ScrollMedium, settings.ScrollSoft} {
			t.Run(fmt.Sprintf("table%d-mode%d", table, scroll), func(t *testing.T) {
				setup := func(mode settings.ScrollMode) (*Runtime, *physics.Game) {
					c := settings.Legacy()
					c.ScrollMode = mode
					r := configured(t, c)
					runKey(t, r, Space)
					runKey(t, r, Key(int(F1)+table-1))
					runKey(t, r, Enter)
					runTicks(t, r, 220)
					r.Model.Session.Release(32, 0)
					runTicks(t, r, 120)
					g, _, _ := gameParts(r.Model.Session)
					g.SetBall(150, 400, 0, 0, false)
					return r, g
				}
				a, ga := setup(settings.ScrollOff)
				b, gb := setup(scroll)
				for tick, tilt := range []bool{true, false, true, false, true} {
					input := physics.Inputs{Tilt: tilt, Left: true, Right: true}
					if err := a.Model.Session.Sync(input); err != nil {
						t.Fatal(err)
					}
					if err := b.Model.Session.Sync(input); err != nil {
						t.Fatal(err)
					}
					if ga.ScreenPosition != gb.ScreenPosition || ga.ScreenSpeed != gb.ScreenSpeed || ga.ScreenOffset != gb.ScreenOffset || ga.TiltCounter != gb.TiltCounter || ga.Tilted != gb.Tilted || ga.AllowFlip != gb.AllowFlip || ga.Ball != gb.Ball {
						t.Fatalf("ScrollOff changed source nudge/tilt/ball at tick %d", tick)
					}
					if tick == 0 && (ga.ScreenPosition != 1800 || ga.ScreenSpeed != 600 || ga.ScreenOffset != 3 || ga.TiltCounter != 60) {
						t.Fatal("OFF skipped original held TILT0 or make-edge TILTLOGIC", ga.ScreenPosition, ga.ScreenOffset, ga.TiltCounter)
					}
					if tick == 2 && (ga.TiltCounter <= 60 || ga.Tilted) {
						t.Fatal("DANGER threshold changed")
					}
					if tick == 4 && (!ga.Tilted || ga.AllowFlip) {
						t.Fatal("TILT threshold/flipper inhibition changed")
					}
				}
				// Rendering the same physical state must expose SCREENPOSY in OFF
				// while retaining exact matrix bytes above the moving table.
				frame := a.Frame()
				offset := ga.ScreenOffset
				ga.ScreenOffset = 0
				zero := a.Frame()
				ga.ScreenOffset = offset
				if offset == 0 || bytes.Equal(frame.Pix, zero.Pix) {
					t.Fatal("OFF nudge invisible")
				}
				if !bytes.Equal(frame.Pix[:33*frame.Stride], zero.Pix[:33*zero.Stride]) {
					t.Fatal("OFF nudge moved matrix")
				}
			})
		}
	}
}
