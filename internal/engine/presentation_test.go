package engine

import (
	"bytes"
	"fmt"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"reflect"
	"testing"
)

// Optional private-original regression. CI skips the loader fixtures explicitly;
// asset-free orientation/settings policy lives in physics/presentation_override_test.
func TestTransientPresentationAllTables(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for scroll := settings.ScrollHard; scroll <= settings.ScrollOff; scroll++ {
			t.Run(fmt.Sprintf("table%d-scroll%d", table, scroll), func(t *testing.T) {
				a, b := New(runtimeFor(t, scroll), 0), New(runtimeFor(t, scroll), 0)
				advance := func(ns int64) {
					var ap, bp []byte
					if err := a.Advance(ns, func(p []byte) error { ap = append(ap, p...); return nil }); err != nil {
						t.Fatal(err)
					}
					if err := b.Advance(ns, func(p []byte) error { bp = append(bp, p...); return nil }); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(ap, bp) || a.State() != b.State() {
						t.Fatal("presentation changed cadence/state/PCM")
					}
				}
				a.Key(57)
				b.Key(57)
				advance(0)
				a.Key(uint8(58 + table))
				b.Key(uint8(58 + table))
				advance(16666667)
				if b.runner.Runtime.Model.Mode != frontend.TableAttract {
					t.Fatal("not attract")
				}
				saved := b.runner.Runtime.Model.Settings
				check := func() {
					b.SetPresentation(true)
					f := b.Frame()
					if f.Rect.Dx() != 320 || f.Rect.Dy() != 609 {
						t.Fatal("portrait", f.Rect)
					}
					if b.runner.Runtime.Model.Settings != saved {
						t.Fatal("saved settings mutated")
					}
					b.SetPresentation(false)
					af, bf := a.Frame(), b.Frame()
					if af.Rect != bf.Rect || !bytes.Equal(af.Pix, bf.Pix) {
						t.Fatal("landscape not restored")
					}
					same(t, reflect.ValueOf(a.runner.Runtime.Model.Session), reflect.ValueOf(b.runner.Runtime.Model.Session), "session")
				}
				check()
				a.Key(28)
				b.Key(28)
				advance(33333334)
				for i := 1; i <= 80; i++ {
					advance(33333334 + int64(i)*14084508)
					check()
				}
				b.SetPresentation(true)
				if err := b.Close(); err != nil {
					t.Fatal(err)
				}
				if b.runner.Runtime.Model.Settings != saved {
					t.Fatal("shutdown mutated settings")
				}
				stored, err := b.runner.Runtime.Model.SettingsStore.Load()
				if err != nil || stored != saved {
					t.Fatal("persisted portrait preference", stored, err)
				}
			})
		}
	}
}

// Exercise the actual Engine -> Runtime -> table -> physics -> composition path
// with invented zero indexed pixels; no loader or commercial assets are used.
func TestSyntheticTransientFrameAndSavedPreference(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for scroll := settings.ScrollHard; scroll <= settings.ScrollOff; scroll++ {
			saved := settings.Config{ScrollMode: scroll, Resolution: 1, Music: 1}
			p := physics.New(&physics.Table{Initial: &assets.InitialTable{Playfield: &assets.Playfield{Indices: make([]byte, 320*576)}}})
			p.Settings = saved
			display := &presentation.Display{}
			var session frontend.Session
			switch table {
			case 1:
				session = &partyland.Game{Physics: p, Display: display}
			case 2:
				session = &speeddevils.Game{Physics: p, Display: display}
			case 3:
				session = &gameshow.Game{Physics: p, Display: display}
			case 4:
				session = &stones.Game{Physics: p, Display: display}
			}
			rt := &frontend.Runtime{Model: &frontend.Model{Mode: frontend.Playing, Selected: table, Session: session, Settings: saved}, View: &frontend.View{}}
			e := New(rt, 100)
			before := e.State()
			next := e.runner.Next
			landscape := append([]byte(nil), e.Frame().Pix...)
			for _, portrait := range []bool{true, false, true, true, false} {
				e.SetPresentation(portrait)
				frame := e.Frame()
				height := saved.RenderHeight() + 33
				if portrait {
					height = 609
				}
				if frame.Rect.Dx() != 320 || frame.Rect.Dy() != height {
					t.Fatal(table, scroll, portrait, frame.Rect)
				}
				if !portrait && !bytes.Equal(frame.Pix, landscape) {
					t.Fatal("landscape changed")
				}
				if p.PresentationFullTable || p.Settings != saved || rt.Model.Settings != saved {
					t.Fatal("override leaked or settings changed")
				}
				if e.State() != before || e.runner.Next != next || e.last != 100 {
					t.Fatal("frame changed source state/time")
				}
			}
		}
	}
}
