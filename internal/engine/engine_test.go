package engine

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/source"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"testing"
	"time"
)

func runtimeFor(t *testing.T, scroll settings.ScrollMode) *frontend.Runtime {
	t.Helper()
	data := os.Getenv("PF_ENGINE_DATA_DIR")
	if data == "" {
		data = "../.."
	}
	testinputs.Require(t, filepath.Join(data, "INTRO.PRG"), filepath.Join(data, "TABLE1.MOD"))
	rt, err := frontend.LoadConfigured(data, nil, &settings.Store{Directory: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	rt.Model.Settings.ScrollMode = scroll
	return rt
}

// Compare complete gameplay/matrix/audio producer checkpoints including private
// task slots. Functions compare code addresses (captures are in the state).
func same(t *testing.T, a, b reflect.Value, path string) {
	t.Helper()
	if a.Type() != b.Type() {
		t.Fatal(path, "type")
	}
	switch a.Kind() {
	case reflect.Pointer, reflect.Interface:
		if a.IsNil() != b.IsNil() {
			t.Fatal(path, "nil")
		}
		if !a.IsNil() {
			same(t, a.Elem(), b.Elem(), path)
		}
	case reflect.Func:
		if a.Pointer() != b.Pointer() {
			t.Fatal(path, "task slot")
		}
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			same(t, a.Field(i), b.Field(i), path+"."+a.Type().Field(i).Name)
		}
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			t.Fatal(path, "length")
		}
		for i := 0; i < a.Len(); i++ {
			same(t, a.Index(i), b.Index(i), path)
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
			same(t, a.MapIndex(k), v, path)
		}
	case reflect.Bool:
		if a.Bool() != b.Bool() {
			t.Fatal(path, a.Bool(), b.Bool())
		}
	case reflect.String:
		if a.String() != b.String() {
			t.Fatal(path, "string")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a.Int() != b.Int() {
			t.Fatal(path, a.Int(), b.Int())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if a.Uint() != b.Uint() {
			t.Fatal(path, a.Uint(), b.Uint())
		}
	case reflect.Float32, reflect.Float64:
		if a.Float() != b.Float() {
			t.Fatal(path, "float")
		}
	default:
		t.Fatalf("unhandled state %s %s", path, a.Kind())
	}
}
func TestDirectRunnerConformanceAllTables(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for _, scroll := range []settings.ScrollMode{settings.ScrollSoft, settings.ScrollOff} {
			t.Run(fmt.Sprintf("table%d-scroll%d", table, scroll), func(t *testing.T) {
				a, b := runtimeFor(t, scroll), runtimeFor(t, scroll)
				now := time.Unix(0, 0)
				direct := source.New(a, func() time.Time { return now })
				e := New(b, 0)
				held := gameplay.Controls{}
				mouse := gameplay.Mouse{}
				delta := 0
				fire := false
				active := func() bool {
					v, ok := a.Model.Session.(interface{ InChute() bool })
					return !direct.Suspended && a.Model.Mode == frontend.Playing && ok && v.InChute()
				}
				key := func(k uint8) {
					direct.Submit(frontend.Input{Gameplay: held, Keys: []frontend.Key{frontend.Key(k)}})
					e.Key(k)
				}
				action := func(k Action, on bool) {
					edge := gameplay.Controls{}
					switch k {
					case Left:
						held.Left = on
					case Right:
						held.Right = on
					case Spring:
						edge.Release = held.Down && !on
						held.Down = on
					case Tilt:
						held.Tilt = on
					}
					edge.Left, edge.Right, edge.Down, edge.Tilt = held.Left, held.Right, held.Down, held.Tilt
					direct.Submit(frontend.Input{Gameplay: edge})
					if err := e.SetAction(k, on); err != nil {
						t.Fatal(err)
					}
				}
				var samples int
				seenLaunch, seenDrain := false, false
				seenPlayers := map[int]bool{}
				seenBalls := map[uint64]bool{}
				check := func(checkpoint bool) {
					var ap, bp []byte
					if err := direct.Advance(func() {
						if !active() {
							mouse.Clear()
							delta = 0
							fire = false
						}
						c := held
						c.MouseY = mouse.Motion(delta)
						c.MouseFire = fire
						delta = 0
						fire = false
						direct.Submit(frontend.Input{Gameplay: c})
					}, func(p []byte) error { ap = append(ap, p...); return nil }); err != nil {
						t.Fatal(err)
					}
					if err := e.Advance(now.UnixNano(), func(p []byte) error { bp = append(bp, p...); return nil }); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(ap, bp) {
						t.Fatal("PCM bytes/sample counts")
					}
					samples += len(ap) / 4
					if a.Model.Session != nil {
						v := reflect.ValueOf(a.Model.Session).Elem()
						g := v.FieldByName("Physics").Interface().(*physics.Game)
						seenLaunch = seenLaunch || !g.SpringValid
						seenDrain = seenDrain || g.Ball.Lost
						if a.Model.Mode == frontend.Playing {
							seenBalls[v.FieldByName("BallNumber").Uint()] = true
							seenPlayers[a.Model.Session.(interface{ CurrentPlayer() int }).CurrentPlayer()] = true
						}
					}

					if a.Model.Tick != b.Model.Tick || a.Model.Mode != b.Model.Mode || a.Model.Selected != b.Model.Selected || direct.Ticks != e.State().Tick || direct.Next != e.runner.Next {
						t.Fatal("source state/deadline parity")
					}
					if checkpoint {
						af, bf := direct.Frame(), e.Frame()
						if af.Rect != bf.Rect || sha256.Sum256(af.Pix) != sha256.Sum256(bf.Pix) {
							t.Fatal("frame parity")
						}
						if a.Model.Session != nil {
							same(t, reflect.ValueOf(a.Model.Session), reflect.ValueOf(b.Model.Session), "session")
						}
						same(t, reflect.ValueOf(a.Model.Scores), reflect.ValueOf(b.Model.Scores), "scores")
					}
				}
				check(true)
				now = direct.Next
				key(uint8(frontend.Space))
				check(true)
				now = direct.Next
				key(uint8(int(frontend.F1) + table - 1))
				check(true)
				// Every restored DOS code, including prefix recovery and message replacement.
				letters := "QWERTYUIOPASDFGHJKLZXCVBNM"
				codes := []uint8{16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 30, 31, 32, 33, 34, 35, 36, 37, 38, 44, 45, 46, 47, 48, 49, 50}
				for _, word := range []string{"EARTHQUAKE", "EXTRA BALLS", "FAIR PLAY", "SNAIL", "CHEAT", "JOHAN", "DANIEL", "GABRIEL", "TSP", "TECH", "ROBBAN", "STEIN", "GREET", "EXTRA BALLS"} {
					for _, c := range word {
						now = direct.Next
						if c == ' ' {
							key(57)
						} else {
							for i, v := range letters {
								if v == c {
									key(codes[i])
									break
								}
							}
						}
						check(false)
					}
					now = direct.Next
					check(true)
				}
				now = direct.Next
				key(uint8(frontend.F2))
				check(true) // two-player game
				if a.Model.Selected != table || a.Model.Mode != frontend.Playing || a.Model.Session.(interface{ PlayerCount() int }).PlayerCount() != 2 {
					t.Fatal("table/hotseat start not exercised")
				}

				for tick := 0; tick < 2400; tick++ {
					now = direct.Next
					action(Left, tick%93 < 18)
					action(Right, tick%71 < 14)
					action(Spring, tick < 32)
					if tick >= 600 && tick%600 < 32 {
						e.PlungerDelta(8)
						if active() {
							delta += 8
						}
					}
					if tick >= 600 && tick%600 == 32 {
						e.PlungerFire()
						if active() {
							fire = true
						}
					}
					if tick == 120 {
						action(Tilt, true)
					}
					if tick == 121 {
						action(Tilt, false)
					}
					if tick == 300 {
						if err := direct.LoseFocus(); err != nil {
							t.Fatal(err)
						}
						if err := e.Suspend(); err != nil {
							t.Fatal(err)
						}
						held = gameplay.Controls{}
						mouse.Clear()
						delta = 0
						fire = false
						now = now.Add(time.Hour)
						check(true)
						direct.Resume()
						if err := e.Resume(now.UnixNano()); err != nil {
							t.Fatal(err)
						}
						check(true)
						if a.Model.Mode != frontend.Paused {
							t.Fatal("focus regain unpaused")
						}
						now = direct.Next
						key(uint8(frontend.P))
						check(true)
						if a.Model.Mode != frontend.Playing {
							t.Fatal("manual resume did not play")
						}
					}
					// Slow presentation wakeup drains all due source ticks and PCM in order.
					if tick%137 == 0 {
						now = now.Add(5 * time.Second / 71)
					}
					check(tick%211 == 0)
				}
				t.Logf("source tasks=%d PCM frames=%d mode=%d launch=%t drain=%t players=%v balls=%v", direct.Ticks, samples, a.Model.Mode, seenLaunch, seenDrain, seenPlayers, seenBalls)
				if !seenLaunch || !seenDrain || len(seenPlayers) < 2 {
					t.Fatal("launch/drain/new-ball/hotseat trace coverage incomplete")
				}

			})
		}
	}
}
func TestClockAndActions(t *testing.T) {
	rt := runtimeFor(t, settings.ScrollSoft)
	e := New(rt, 10)
	if e.Advance(9, nil) == nil {
		t.Fatal("backward clock accepted")
	}
	if e.SetAction(Action(99), true) == nil {
		t.Fatal("unknown action accepted")
	}
	e.PlungerDelta(7)
	if e.delta != 0 {
		t.Fatal("invalid spring banked movement")
	}
}

func TestShutdownUsesSharedSettingsStorage(t *testing.T) {
	data := os.Getenv("PF_ENGINE_DATA_DIR")
	if data == "" {
		data = "../.."
	}
	testinputs.Require(t, filepath.Join(data, "INTRO.PRG"))
	state := t.TempDir()
	e, err := Load(data, state, 0)
	if err != nil {
		t.Fatal(err)
	}
	// The existing settings model owns encoding and persistence. Host destruction
	// must execute frontend shutdown even while lifecycle-suspended.
	e.runner.Runtime.Model.Settings.ScrollMode = settings.ScrollOff
	if err := e.Suspend(); err != nil {
		t.Fatal(err)
	}
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}
	c, err := (settings.Store{Directory: state}).Load()
	if err != nil || c.ScrollMode != settings.ScrollOff {
		t.Fatal("shutdown persistence", c, err)
	}
	if !e.State().Done {
		t.Fatal("shutdown did not quit")
	}
}
