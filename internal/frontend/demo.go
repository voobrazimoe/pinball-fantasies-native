package frontend

import (
	"errors"
	"fmt"
	"image"
	"io/fs"
	"os"
	"path/filepath"

	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
)

// DemoNamesRequired is the official 10-minute demo's runtime fingerprint.
// Launcher, drivers, TIMER.BIN and settings files are not runtime roles.
var DemoNamesRequired = []string{"INTRO.PRG", "INTRO.MOD", "MOD2.MOD", "TABLE1.PRG", "TABLE1.MOD"}

// DemoProfileID marks a runtime loaded from the official 10-minute demo.
const DemoProfileID = datalayout.PartyLandDemoProfile

// loadDemo enters Party Land directly. The demo INTRO selector/attract is not
// reproduced: TABLE2-4 are absent and its advertising cards are not tables.
func loadDemo(dataDir string) (*Runtime, error) {
	inputs := make(map[string][]byte)
	for _, name := range DemoNamesRequired {
		data, err := os.ReadFile(filepath.Join(dataDir, name))
		if err != nil {
			return nil, err
		}
		inputs[name] = data
	}
	id, err := datalayout.DetectDemoInstallation(inputs["INTRO.PRG"], inputs["TABLE1.PRG"])
	if err != nil {
		return nil, err
	}
	data, err := datalayout.PreparePRGForProfile(id, "TABLE1.PRG", inputs["TABLE1.PRG"])
	if err != nil {
		return nil, err
	}
	records, err := datalayout.DemoRecords(inputs["TABLE1.PRG"])
	if err != nil {
		return nil, err
	}
	table, err := physics.DecodePartyLand(data)
	if err != nil {
		return nil, err
	}
	module, err := audio.Decode(inputs["TABLE1.MOD"])
	if err != nil {
		return nil, fmt.Errorf("TABLE1.MOD: %w", err)
	}
	// Role-specific decoders keep the five-file installation coherent.
	if _, err = audio.DecodeIntro(inputs["INTRO.MOD"]); err != nil {
		return nil, fmt.Errorf("INTRO.MOD: %w", err)
	}
	if _, err = audio.DecodeMenu(inputs["MOD2.MOD"]); err != nil {
		return nil, fmt.Errorf("MOD2.MOD: %w", err)
	}
	s := &demoSession{newGame: func() *partyland.Game {
		g := partyland.NewDemo(table, data, partyland.DemoInputs{PlayersText: records.PlayersText, ExpiryTexts: records.ExpiryTexts})
		// The proved reference uses the accepted Legacy native settings.
		g.Configure(settings.Legacy())
		// High scores are volatile factory values: no HI load/save is proved.
		g.SetHighScore(Defaults(1)[0].Digits)
		g.AttachAudio(module)
		return g
	}}
	r := &Runtime{ProfileID: id, Model: &Model{Mode: Playing, Selected: 1}}
	s.runtime = r
	s.start()
	r.session = s
	return r, nil
}

// A folder without any of TABLE2-4.PRG is offered to the demo detector; a
// partial full installation keeps its ordinary missing-file error.
func missingFullInstallation(err error) bool { return errors.Is(err, fs.ErrNotExist) }
func demoFolder(dataDir string) bool {
	for _, name := range []string{"TABLE2.PRG", "TABLE3.PRG", "TABLE4.PRG"} {
		if _, err := os.Stat(filepath.Join(dataDir, name)); !errors.Is(err, fs.ErrNotExist) {
			return false
		}
	}
	return true
}

// demoSession is the native lifecycle around one table lifetime: P pauses
// (timer, expired flag and matrix are untouched), Esc asks to quit, Y quits.
// The demo's linked QUIT ends the program as the DOS demo exits to DOS.
type demoSession struct {
	runtime *Runtime
	game    *partyland.Game
	newGame func() *partyland.Game
}

func (s *demoSession) start() {
	s.game = s.newGame()
	m := s.runtime.Model
	m.Session, m.Mode, m.Selected, m.PauseDelay = s.game, Playing, 1, 0
}

func (s *demoSession) FocusLost(close bool) error {
	m := s.runtime.Model
	s.runtime.PCM = nil
	if close {
		m.Mode = Quit
	} else if m.Mode == Playing {
		m.Mode = Paused
	}
	return nil
}

func (s *demoSession) Update(in Input) error {
	r, m := s.runtime, s.runtime.Model
	if in.Close {
		m.Mode = Quit
		return nil
	}
	if in.FocusLost {
		return s.FocusLost(false)
	}
	for _, k := range in.Keys {
		before := m.Mode
		switch m.Mode {
		case Playing:
			switch {
			case k == P && m.PauseDelay == 0:
				m.Mode = Paused
			case k == Escape:
				m.ReturnMode, m.Mode = Playing, QuitQuestion
			case k == 50:
				s.game.ToggleMusic()
			}
		case Paused:
			if k == Escape {
				m.ReturnMode, m.Mode = Playing, QuitQuestion
			} else {
				m.Mode, m.PauseDelay = Playing, 30
			}
		case QuitQuestion:
			if Initial(k) == 'Y' {
				m.End = Aborted
				m.Mode = Quit
			} else {
				m.Mode, m.PauseDelay = m.ReturnMode, 30
			}
		}
		if m.Mode != before {
			return nil // transition keys never reach the table
		}
	}
	if m.Mode != Playing {
		return nil
	}
	if m.PauseDelay > 0 {
		m.PauseDelay--
	}
	if err := s.game.Sync(in.controls()); err != nil {
		return err
	}
	r.PCM = s.game.PCM()
	if s.game.DemoFinished() {
		m.End = ProgramQuit
		m.Mode = Quit
	} else if _, over := s.game.Result(); over {
		// Only a canonical fallback (match/extra-ball tail) can end the game.
		m.End = Completed
		m.Mode = Quit
	}
	return nil
}

func (s *demoSession) Frame(full bool) *image.RGBA {
	p := s.game.Physics
	previous := p.PresentationFullTable
	p.PresentationFullTable = full
	defer func() { p.PresentationFullTable = previous }()
	frame := s.game.Frame()
	if s.runtime.Model.Suspended() {
		// Native pause/quit feedback; the table state is not touched.
		out := image.NewRGBA(frame.Rect)
		for i, v := range frame.Pix {
			if i%4 == 3 {
				out.Pix[i] = v
			} else {
				out.Pix[i] = v / 2
			}
		}
		return out
	}
	return frame
}

func (s *demoSession) Diagnostic() string { return "" }
