// Package source owns deterministic progression, with no window or device API.
package source

import (
	"image"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gameplay"
	"time"
)

// Runner advances every due source tick. Now must return monotonic source time.
// A PCM sink is called for each tick before returning to host presentation.
type Runner struct {
	Runtime          *frontend.Runtime
	Now              func() time.Time
	Next             time.Time
	Done, Suspended  bool
	input            frontend.Input
	resumeAfterFocus bool
	mouseFireNext    bool
}

func New(r *frontend.Runtime, now func() time.Time) *Runner {
	return &Runner{Runtime: r, Now: now, Next: now()}
}
func (r *Runner) Frame() *image.RGBA { return r.Runtime.Frame() }

// Submit replaces held state and accumulates edges until a source tick consumes
// them. Hosts may poll between ticks without losing short presses or motion.
func (r *Runner) Submit(in frontend.Input) {
	if in.FocusLost {
		r.input = frontend.Input{FocusLost: true, Close: in.Close || r.input.Close}
		r.mouseFireNext = false
		return
	}
	if r.Suspended {
		if in.Close {
			r.input.Close = true
		}
		return
	}
	r.input.Gameplay.Left = in.Gameplay.Left
	r.input.Gameplay.Right = in.Gameplay.Right
	r.input.Gameplay.Down = in.Gameplay.Down
	r.input.Gameplay.Tilt = in.Gameplay.Tilt
	r.input.Gameplay.Release = r.input.Gameplay.Release || in.Gameplay.Release
	r.input.Gameplay.MouseFire = r.input.Gameplay.MouseFire || in.Gameplay.MouseFire
	r.input.Gameplay.MouseY += in.Gameplay.MouseY
	r.input.Keys = append(r.input.Keys, in.Keys...)
	r.input.Close = r.input.Close || in.Close
}

// Suspend releases controls without touching game or tracker state. Desktop
// focus loss separately submits FocusLost to retain its existing pause request.
func (r *Runner) Suspend() {
	r.Suspended = true
	r.input = frontend.Input{}
	r.mouseFireNext = false
	r.resumeAfterFocus = false
}
func (r *Runner) Resume() {
	if r.input.FocusLost {
		r.resumeAfterFocus = true
		return
	}
	if !r.Suspended {
		return
	}
	r.Suspended = false
	r.input = frontend.Input{}
	r.Next = r.Now()
}

// Advance polls input per due tick and submits all PCM, even after a slow host
// frame. Cadence uses the mode BEFORE Update, exactly like the DOS host loop.
func (r *Runner) Advance(poll func(), pcm func([]byte) error) error {
	if r.Done {
		return nil
	}
	if r.Suspended {
		if r.input.Close {
			if err := r.Runtime.Update(frontend.Input{Close: true}); err != nil {
				return err
			}
			r.Done = true
		}
		return nil
	}
	for !r.Now().Before(r.Next) {
		if poll != nil {
			poll()
		}
		if r.Suspended || r.Done {
			return nil
		}
		hz := r.Runtime.Model.Hz()
		in := r.input
		// SPRINGSTEEN installs SPRINGUP for the following source task, after its
		// mouse adjustment and SPRINGIT have run. Keep this edge out of host time.
		fire := in.Gameplay.MouseFire
		in.Gameplay.MouseFire = r.mouseFireNext
		r.mouseFireNext = fire
		if in.FocusLost {
			in = frontend.Input{FocusLost: true, Close: in.Close}
		}
		held := in.Gameplay
		r.input = frontend.Input{Gameplay: gameplay.Controls{Left: held.Left, Right: held.Right, Down: held.Down, Tilt: held.Tilt}}
		if in.FocusLost {
			r.input = frontend.Input{}
		}
		if in.FocusLost {
			// Apply the desktop pause request without rendering INTRO PCM or moving
			// tracker phase just because a lifecycle notification arrived.
			r.Runtime.PCM = nil
			if err := r.Runtime.Model.Update(in); err != nil {
				return err
			}
		} else if err := r.Runtime.Update(in); err != nil {
			return err
		}
		r.Next = r.Next.Add(time.Second / time.Duration(hz))
		r.Done = r.Runtime.Model.Mode == frontend.Quit
		if in.FocusLost {
			if r.resumeAfterFocus {
				r.input = frontend.Input{}
				r.Next = r.Now().Add(time.Second / time.Duration(hz))
				r.resumeAfterFocus = false
			} else {
				r.Suspend()
			}
		}
		if pcm != nil {
			if err := pcm(r.Runtime.PCM); err != nil {
				return err
			}
		}
		if r.Done || r.Suspended {
			return nil
		}
	}
	return nil
}
