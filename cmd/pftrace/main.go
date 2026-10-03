// pftrace is a deterministic direct-Runner oracle for the foreign C ABI harness.
// It reads a JSON host trace from stdin and never embeds commercial inputs.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"hash"
	"os"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/source"
	"time"
)

type Step struct {
	NS                           int64
	Keys                         []uint8
	Actions                      [][2]int
	Delta                        int
	Fire, Suspend, Resume, Frame bool
}
type Trace struct{ Steps []Step }
type Result struct {
	Tick                  uint64
	Mode, Table, Flags    uint32
	PCM                   string
	Frames                int
	Frame                 string
	Width, Height, Stride int
}

func main() {
	data := flag.String("data", "", "original data directory")
	state := flag.String("state", "", "private state directory")
	flag.Parse()
	var trace Trace
	must(json.NewDecoder(os.Stdin).Decode(&trace))
	if len(trace.Steps) == 0 {
		panic("empty trace")
	}
	rt, err := frontend.LoadConfigured(*data, frontend.FileStore{Directory: *state, SeedDirectory: *data}, &settings.Store{Directory: *state, SeedDirectory: *data})
	must(err)
	now := time.Unix(0, trace.Steps[0].NS)
	r := source.New(rt, func() time.Time { return now })
	held := gameplay.Controls{}
	mouse := gameplay.Mouse{}
	delta := 0
	fire := false
	active := func() bool {
		s, ok := rt.Model.Session.(interface{ InChute() bool })
		return !r.Suspended && rt.Model.Mode == frontend.Playing && ok && s.InChute()
	}
	submit := func(c gameplay.Controls) {
		c.Left, c.Right, c.Down, c.Tilt = held.Left, held.Right, held.Down, held.Tilt
		r.Submit(frontend.Input{Gameplay: c})
	}
	results := make([]Result, 0, len(trace.Steps))
	var pcm hash.Hash
	for _, step := range trace.Steps {
		now = time.Unix(0, step.NS)
		if step.Suspend {
			must(r.LoseFocus())
			held = gameplay.Controls{}
			mouse.Clear()
			delta = 0
			fire = false
		}
		if step.Resume {
			if r.Suspended {
				held = gameplay.Controls{}
				mouse.Clear()
				delta = 0
				fire = false
				r.Resume()
			}
		}
		for _, a := range step.Actions {
			if r.Suspended {
				continue
			}
			on := a[1] != 0
			c := gameplay.Controls{}
			switch a[0] {
			case 0:
				held.Left = on
			case 1:
				held.Right = on
			case 2:
				c.Release = held.Down && !on
				held.Down = on
			case 3:
				held.Tilt = on
			default:
				panic("bad action")
			}
			submit(c)
		}
		for _, key := range step.Keys {
			r.Submit(frontend.Input{Gameplay: held, Keys: []frontend.Key{frontend.Key(key)}})
		}
		if active() {
			delta += step.Delta
			fire = fire || step.Fire
		} else {
			mouse.Clear()
			delta = 0
			fire = false
		}
		pcm = sha256.New()
		frames := 0
		must(r.Advance(func() {
			if !active() {
				mouse.Clear()
				delta = 0
				fire = false
			}
			submit(gameplay.Controls{MouseY: mouse.Motion(delta), MouseFire: fire})
			delta = 0
			fire = false
		}, func(p []byte) error { pcm.Write(p); frames += len(p) / 4; return nil }))
		flags := uint32(0)
		if r.Suspended {
			flags |= 1
		}
		if r.Done {
			flags |= 2
		}
		if active() {
			flags |= 4
		}
		out := Result{Tick: r.Ticks, Mode: uint32(rt.Model.Mode), Table: uint32(rt.Model.Selected), Flags: flags, PCM: hex.EncodeToString(pcm.Sum(nil)), Frames: frames}
		if step.Frame {
			frame := r.Frame()
			h := sha256.Sum256(frame.Pix)
			out.Frame = fmt.Sprintf("%x", h)
			out.Width = frame.Rect.Dx()
			out.Height = frame.Rect.Dy()
			out.Stride = frame.Stride
		}
		results = append(results, out)
	}
	must(json.NewEncoder(os.Stdout).Encode(results))
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
