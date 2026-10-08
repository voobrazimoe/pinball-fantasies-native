//go:build demodev

package frontend

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
)

type experimentalSession struct {
	runtime    *Runtime
	game       *partyland.ExperimentalDemo
	diagnostic string
	logPath    string
}

func LoadExperimentalDemo(demo, canonical, logPath string) (*Runtime, error) {
	g, err := partyland.LoadExperimentalDemo(demo, canonical)
	if err != nil {
		return nil, err
	}
	// Prove diagnostics are writable before opening gameplay. This is separate
	// from the original installation and contains no original file bytes.
	for _, dir := range []string{demo, canonical} {
		a, _ := filepath.Abs(dir)
		b, _ := filepath.Abs(logPath)
		if filepath.Dir(b) == a {
			return nil, fmt.Errorf("demo log must be outside original input directories")
		}
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	if err = f.Close(); err != nil {
		return nil, err
	}
	r := &Runtime{Model: &Model{Mode: Playing, Selected: 1}, ProfileID: "experimental-10min-demo"}
	r.session = &experimentalSession{runtime: r, game: g, logPath: logPath}
	return r, nil
}
func (s *experimentalSession) FocusLost(close bool) error {
	s.runtime.PCM = nil
	if close {
		s.runtime.Model.Mode = Quit
	} else if s.runtime.Model.Mode == Playing {
		s.runtime.Model.Mode = Paused
	}
	return nil
}
func (s *experimentalSession) Update(in Input) error {
	r := s.runtime
	if in.Close {
		r.Model.Mode = Quit
		return nil
	}
	if in.FocusLost {
		return s.FocusLost(false)
	}
	for _, k := range in.Keys {
		if k == Escape {
			r.Model.Mode = Quit
			return nil
		}
		if k == P {
			if r.Model.Mode == Paused && s.diagnostic == "" {
				r.Model.Mode = Playing
			} else {
				r.Model.Mode = Paused
			}
		} else if r.Model.Mode == Paused && s.diagnostic == "" {
			r.Model.Mode = Playing
		}
	}
	if r.Model.Mode != Playing || s.diagnostic != "" {
		return nil
	}
	c := in.controls()
	err := s.game.Sync(physics.Inputs{Left: c.Left, Right: c.Right, Down: c.Down, Release: c.Release, Tilt: c.Tilt})
	if err != nil {
		s.diagnostic = s.game.Diagnostic()
		if s.diagnostic == "" {
			s.diagnostic = err.Error()
		}
		r.Model.Mode = Paused
		f, e := os.OpenFile(s.logPath, os.O_WRONLY|os.O_APPEND, 0600)
		if e == nil {
			_, e = fmt.Fprintln(f, s.diagnostic)
			ce := f.Close()
			if e == nil {
				e = ce
			}
		}
		if e != nil {
			s.diagnostic += "\nDiagnostic log write failed: " + e.Error()
		}
		fmt.Fprintln(os.Stderr, s.diagnostic)
		return nil // frozen, still accepts Escape/window close
	}
	r.PCM = s.game.PCM()
	if s.game.Done() {
		r.Model.Mode = Quit
	}
	return nil
}
func (s *experimentalSession) Frame(full bool) *image.RGBA { return s.game.Frame() }
func (s *experimentalSession) Diagnostic() string          { return s.diagnostic }
