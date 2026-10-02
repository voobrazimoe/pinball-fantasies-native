// Package frontend owns INTRO's selector and FANTASIE's single-player lifecycle.
// It calls a table session at whole deterministic sync boundaries.
package frontend

import (
	"image"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/tablelogic"
)

type Mode uint8

const (
	Startup Mode = iota
	Selector
	SelectorText
	Options
	TableAttract
	Playing
	Paused
	QuitQuestion
	GameEnd
	Initials
	EntryWait
	Quit
)

type Reason uint8

const (
	Active Reason = iota
	Completed
	Aborted
	ProgramQuit
)

// Keys are DOS make codes. Releases/held flippers are separate input, avoiding
// host key repeat as a game-visible clock.
type Key uint8

const (
	Escape Key = 1
	Enter  Key = 28
	P      Key = 25
	Space  Key = 57
	F1     Key = 59
	F2     Key = 60
	F3     Key = 61
	F4     Key = 62
	F5     Key = 63
	F6     Key = 64
	F7     Key = 65
	F8     Key = 66
	Up     Key = 72
	Down   Key = 80
)

type Input struct {
	Gameplay                                gameplay.Controls
	Keys                                    []Key
	Left, Right, Down, Release, Tilt, Close bool
	FocusLost                               bool
}
type Session interface {
	Sync(physics.Inputs) error
	Release(uint8, uint8)
	Frame() *image.RGBA
	Result() (tablelogic.Decimal, bool)
	PCM() []byte
	Cue(string)
}
type Factory func(top tablelogic.Decimal) (Session, error)
type Model struct {
	scoreQueue      []tablelogic.Decimal
	highscorePlayed bool
	scorePlayer     int
	selectionOpen   bool
	selectionDelay  int
	OptionsPending  bool
	OptionsOrigin   Mode

	Settings                             settings.Config
	SettingsStore                        *settings.Store
	OptionRow, OptionTick                int
	OptionsLeaving                       bool
	Mode                                 Mode
	Selected                             int
	Page                                 int
	Counter                              int
	Tick                                 uint64
	End                                  Reason
	Scores                               [4]Scores
	Session                              Session
	Factories                            [4]Factory
	Factory                              Factory
	Store                                Store
	Entry                                [3]byte
	Entered, Rank                        int
	Final                                tablelogic.Decimal
	Segment, SegmentTick                 int
	IntroClock                           uint64
	TextPage                             int
	Reveal                               int
	SidebarTick                          int
	TextTick, TextExitTick, PreviousPage int
	ReturnMode                           Mode
	PauseDelay                           int
	audioMode                            Mode
}

func New(store Store, factory Factory) (*Model, error) {
	m := &Model{Mode: Startup, Settings: settings.Defaults(), Selected: 1, Store: store, Factory: factory, Counter: 540, Rank: -1}
	for i := range m.Scores {
		var e error
		if store == nil {
			m.Scores[i] = Defaults(i + 1)
		} else {
			m.Scores[i], e = store.Load(i + 1)
		}
		if e != nil {
			return nil, e
		}
	}
	return m, nil
}

func (m *Model) tableFactory() Factory {
	if m.Factories[m.Selected-1] != nil {
		return m.Factories[m.Selected-1]
	}
	if m.Selected == 1 {
		return m.Factory
	}
	return nil
}

// INTRO synca uses 60 Hz presentation; Party Land retains PF4.5's 71 Hz sync.
func (m *Model) Hz() int {
	if m.Mode >= TableAttract && m.Mode < Quit {
		return 71
	}
	return 60
}
func (m *Model) Suspended() bool { return m.Mode == Paused || m.Mode == QuitQuestion }
func (m *Model) selector() {
	m.Mode = Selector
	m.Counter = 540
	m.Session = nil
	m.Reveal = 0
	m.SidebarTick = 0
}
func (m *Model) start() error { return m.startPlayers(1) }
func (m *Model) startPlayers(count int) error {
	s, e := m.tableFactory()(m.Scores[m.Selected-1][0].Digits)
	if e != nil {
		return e
	}
	if g, ok := s.(interface{ CarryLoadedTableState(any) }); ok {
		g.CarryLoadedTableState(m.Session)
	}
	m.Session = s
	if g, ok := s.(interface{ StartPlayers(int) }); ok {
		g.StartPlayers(count)
	}
	m.scoreQueue = nil
	m.highscorePlayed = false
	m.scorePlayer = 1
	m.selectionOpen, m.selectionDelay = true, 15
	m.Mode = Playing
	m.End = Active
	return nil
}
func (m *Model) loadTable(table int) error {
	if e := m.saveSettings(); e != nil {
		return e
	}
	m.Selected = table
	m.Session = nil
	if m.tableFactory() == nil {
		return nil
	} // Original choice is retained; unavailable program never loads another table.
	s, e := m.tableFactory()(m.Scores[m.Selected-1][0].Digits)
	if e != nil {
		return e
	}
	m.Session = s
	m.Mode = TableAttract
	m.End = Active
	m.Final = tablelogic.Decimal{}
	m.Counter = 0
	if table, ok := s.(interface{ AttractCue() string }); ok {
		s.Cue(table.AttractCue())
	} else {
		s.Cue("S_MAIN")
	}
	return nil
}
func (m *Model) finish(d tablelogic.Decimal) {
	m.End = Completed
	m.Final = d
	m.Rank = m.Scores[m.Selected-1].Rank(d)
	m.Entry = [3]byte{' ', ' ', ' '}
	m.Entered = 0
	if m.Rank >= 0 {
		m.Mode = Initials
		if !m.highscorePlayed {
			m.Session.Cue("S_GAMEOVER2")
			m.highscorePlayed = true
		}
	} else {
		m.attract()
		m.Session.Cue("S_GAMEOVER")
	}
}
func (m *Model) attract() {
	m.Mode = TableAttract
	m.Counter = 0
	if s, ok := m.Session.(interface{ AttractCue() string }); ok {
		m.Session.Cue(s.AttractCue())
	}
}

// ALFA_KEYS from linked TABLE1.PRG at 1d396: letters and Space as '*'.
func Initial(k Key) byte {
	const keys = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00QWERTYUIOP\x00\x00\x00\x00ASDFGHJKL\x00\x00\x00\x00\x00ZXCVBNM\x00\x00\x00\x00\x00\x00*"
	if int(k) < len(keys) {
		return keys[k]
	}
	return 0
}
func (m *Model) Update(in Input) error {
	m.Tick++
	if in.Close {
		if e := m.saveSettings(); e != nil {
			return e
		}
		m.Mode = Quit
		m.End = ProgramQuit
		return nil
	}
	// Host focus loss is a pause request, never a toggle. It takes precedence
	// over queued makes/releases and ignores PauseDelay from a recent resume.
	if in.FocusLost {
		if m.Mode == Playing {
			m.Mode = Paused
		}
		return nil
	}
	initialMode := m.Mode
	for _, k := range in.Keys {
		keyMode := m.Mode
		switch m.Mode {
		case Startup:
			if k == Space {
				m.selector()
			}
		case Selector, SelectorText:
			if k == Escape {
				if e := m.saveSettings(); e != nil {
					return e
				}
				m.Mode = Quit
				m.End = ProgramQuit
				return nil
			}
			if k == F5 {
				if m.Mode == SelectorText {
					if m.TextTick >= 45 {
						m.OptionsPending = true
						m.Counter = 0
						m.TextExitTick = 0
					}
					break
				}
				m.openOptions()
				break
			}
			if k >= F1 && k <= F4 {
				if e := m.loadTable(int(k-F1) + 1); e != nil {
					return e
				}
				break
			}
			if k == Space || k == Enter {
				if m.Mode == SelectorText {
					// INTRO reads these make codes only in wait_sync2, after
					// fade3. Skipping its hold still executes fade3b20.
					if m.TextTick >= 45 && m.Counter > 0 {
						m.Counter, m.TextExitTick = 0, 0
						if k == Enter {
							m.Page ^= 1
						}
						return nil // First fade3b frame uses CX=0 (19/20).
					}
				} else {
					m.advanceSelector(k == Enter)
				}
			}
		case Options:
			if m.OptionTick < 45 || m.OptionsLeaving {
				continue
			}
			switch k {
			case Up:
				m.OptionRow = (m.OptionRow + 5) % 6
			case Down:
				m.OptionRow = (m.OptionRow + 1) % 6
			case Enter, Space:
				if m.OptionRow == 5 {
					m.OptionsLeaving = true
					m.OptionTick = 0
				} else {
					m.Settings.Cycle(m.OptionRow)
				}
			case Escape:
				m.OptionsLeaving = true
				m.OptionTick = 0
			}
		case TableAttract:
			if k == Escape {
				m.ReturnMode = TableAttract
				m.Mode = QuitQuestion
			} else if (k >= F1 && k <= F8) || k == Enter {
				// checkstartkeys compares PLAYERS with eight before resetting
				// PLAYERS_CP. After an eight-player game Enter is ignored.
				if k == Enter {
					if g, ok := m.Session.(interface{ PlayerCount() int }); ok && g.PlayerCount() == 8 {
						continue
					}
				}
				count := 1
				if k != Enter {
					count = int(k-F1) + 1
				}
				if e := m.startPlayers(count); e != nil {
					return e
				}
			}
		case Playing:
			if m.selectionOpen && m.selectionDelay == 0 && ((k >= F1 && k <= F8) || k == Enter) {
				if g, ok := m.Session.(interface {
					SelectPlayers(int)
					PlayerCount() int
				}); ok {
					count := int(k-F1) + 1
					if k == Enter {
						count = g.PlayerCount() + 1
					}
					if count <= 8 {
						g.SelectPlayers(count)
						m.selectionDelay = 15
					}
				}
			}
			if k == 50 {
				if g, ok := m.Session.(interface{ ToggleMusic() }); ok {
					g.ToggleMusic()
				}
			}
			if k == P && m.PauseDelay == 0 {
				m.Mode = Paused
			} else if k == Escape && m.SessionReady() {
				m.End = Aborted
				m.attract()
			}
		case Paused:
			if k == Escape {
				m.ReturnMode = Playing
				m.Mode = QuitQuestion
			} else {
				m.Mode = Playing
				m.PauseDelay = 30
			}
		case QuitQuestion:
			if m.ReturnMode == TableAttract && Initial(k) == 0 {
				continue
			}
			if Initial(k) == 'Y' {
				m.End = Aborted
				m.selector()
			} else {
				m.Mode = m.ReturnMode
				m.PauseDelay = 30
			}
		case Initials:
			if c := Initial(k); c != 0 {
				m.Entry[m.Entered] = c
				m.Entered++
				if m.Entered == 3 {
					scores := m.Scores[m.Selected-1]
					scores.Insert(m.Rank, m.Final, m.Entry)
					if m.Store != nil {
						if e := m.Store.Save(m.Selected, scores); e != nil {
							return e
						}
					}
					m.Scores[m.Selected-1] = scores
					m.Mode = EntryWait
					m.Counter = 60
				}
			}
		}
		if m.Mode != keyMode {
			break
		}
	}
	if m.Mode != initialMode {
		return nil
	} // Transition keys never leak into the destination session.
	switch m.Mode {
	case Startup:
		m.IntroClock++
		m.advanceStartup()
	case Selector:
		m.SidebarTick++
		if m.Reveal < 20+len(rasterGroups) {
			m.Reveal++
		} else {
			m.Counter--
			if m.Counter <= 0 {
				m.advanceSelector(false)
			}
		}
	case SelectorText:
		m.SidebarTick++
		if m.TextTick < 45 {
			m.TextTick++
		} else if m.Counter > 0 {
			m.Counter--
		} else if m.TextExitTick < 20 {
			m.TextExitTick++
		} else {
			m.advanceSelector(false)
		}

	case Options:
		m.SidebarTick++
		m.OptionTick++
		if m.OptionsLeaving && m.OptionTick >= 40 {
			m.Mode = Selector
			m.Counter = 540
			m.Reveal = 0
		}
	case Playing:
		if m.selectionDelay > 0 {
			m.selectionDelay--
		}
		ready := m.SessionReady()
		if g, ok := m.Session.(interface{ PlayerSelectionReady() bool }); ok {
			ready = g.PlayerSelectionReady()
		}
		if !ready {
			m.selectionOpen = false
		}
		if m.PauseDelay > 0 {
			m.PauseDelay--
		}
		if e := m.Session.Sync(in.controls()); e != nil {
			return e
		}
		if d, done := m.Session.Result(); done {
			m.Final = d
			m.End = Completed
			m.Mode = GameEnd
			m.Counter = 5
		}
	case GameEnd:
		m.Counter--
		if m.Counter == 0 {
			if s, ok := m.Session.(interface{ PlayerScores() []tablelogic.Decimal }); ok {
				m.scoreQueue = s.PlayerScores()
			} else {
				m.scoreQueue = []tablelogic.Decimal{m.Final}
			}
			m.scorePlayer = 0
			m.highscorePlayed = false
			m.nextScore()
		}
	case TableAttract:
		m.Counter++
	case EntryWait:
		m.Counter--
		if m.Counter <= 2 {
			m.nextScore()
		}
	}
	return nil
}

// SessionReady is the narrow lifecycle capability for Escape in the chute.
func (m *Model) SessionReady() bool {
	s, ok := m.Session.(interface{ InChute() bool })
	return ok && s.InChute()
}
func (m *Model) advanceSelector(enter bool) {
	if m.Mode == SelectorText {
		if m.OptionsPending {
			m.OptionsPending = false
			m.openOptions()
			return
		}
		m.Mode = Selector
		m.Counter = 540
		m.Reveal = 0
		return
	}
	m.PreviousPage = m.Page
	m.Page ^= 1
	// INTRO/TEXTLISTA has ten entries after its dynamic HITEXT slot.
	// TEXTPEK starts at slot1 and wraps at the following zero terminator.
	m.TextPage = m.TextPage%10 + 1
	if enter {
		m.Mode = Selector
		m.Counter = 540
		m.Reveal = 0
		return
	}
	m.Mode = SelectorText
	m.Counter = 420
	m.TextTick = 0
	m.TextExitTick = 0
}

// INTRO wait_music uses a 50 Hz music counter; palette fades use synca.
type startupPart struct{ picture, frames, until, from, to int }

var startupParts = []startupPart{
	{0, 0, 1, 0, 0}, {0, 20, 0, 0, 63}, {0, 0, 302, 63, 63}, {0, 20, 0, 63, 0},
	{1, 10, 0, 127, 63}, {1, 0, 620, 63, 63}, {1, 20, 0, 63, 0},
	{2, 10, 0, 127, 63}, {2, 0, 891, 63, 63}, {2, 20, 0, 63, 0},
	{3, 8, 0, 0, 63}, {3, 0, 0, 63, 63}, {3, 20, 0, 63, 0},
	{4, 20, 0, 0, 63}, {4, 0, 1500, 63, 63}, {4, 20, 0, 63, 0},
}

func (m *Model) advanceStartup() {
	m.SegmentTick++
	p := startupParts[m.Segment]
	if (p.frames > 0 && m.SegmentTick >= p.frames) || (p.frames == 0 && m.IntroClock*50/60 > uint64(p.until)) {
		m.Segment++
		m.SegmentTick = 0
		if m.Segment == len(startupParts) {
			m.selector()
		}
	}
}

func (m Mode) String() string {
	names := [...]string{"startup", "selector", "selector text", "options", "table attract", "playing", "paused", "quit question", "game over", "initials", "entry wait", "quit"}
	if int(m) >= len(names) {
		return "unknown"
	}
	return names[m]
}

func (m *Model) saveSettings() error {
	if m.SettingsStore != nil {
		return m.SettingsStore.Save(m.Settings)
	}
	return nil
}

func (m *Model) openOptions() {
	m.OptionsOrigin = m.Mode
	m.Mode = Options
	m.OptionRow = 0
	m.OptionTick = 0
	m.OptionsLeaving = false
}

// SessionConfig snapshots persistent settings while preserving MUSIC_TOGGLE
// across another game within the loaded table program. Reload from INTRO has
// no session and restores the persisted S_IM byte.
func (m *Model) SessionConfig() settings.Config {
	c := m.Settings
	if live, ok := m.Session.(interface{ MusicDisabled() bool }); ok {
		c.Music = 0
		if live.MusicDisabled() {
			c.Music = 1
		}
	}
	return c
}

// DOS SPINTSEL_IN_HIGH visits PLAYER_AREA in player order and inserts against
// the updated global four-entry list. Equal scores do not displace an equal row.
func (m *Model) nextScore() {
	for m.scorePlayer < len(m.scoreQueue) {
		d := m.scoreQueue[m.scorePlayer]
		m.scorePlayer++
		if m.Scores[m.Selected-1].Rank(d) >= 0 {
			m.finish(d)
			return
		}
	}
	m.attract()
	if !m.highscorePlayed {
		m.Session.Cue("S_GAMEOVER")
	}
}
