// Package frontend owns INTRO's selector and FANTASIE's single-player lifecycle.
// It calls a table session at whole deterministic sync boundaries.
package frontend

import (
	"image"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
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
	// Closing is the demo launcher's exit text after INTRO quits (Esc).
	Closing
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
	sessionSynced                    bool
	sourceScoreStage                 int // 1: SPINTSEL/get/read/wait; 2: source completion tasks.
	entrySetup                       bool
	gameOverTimeline                 *presentation.Timeline
	cheatDecoder                     gameplay.CheatDecoder
	cheatTimeline                    *presentation.CheatTimeline
	cheatTiltDisabled, cheatFastBall bool
	cheatBalls                       int
	scoreQueue                       []tablelogic.Decimal
	highscorePlayed                  bool
	scorePlayer                      int
	selectionOpen                    bool
	selectionDelay                   int
	OptionsPending                   bool
	OptionsOrigin                    Mode

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
	// Demo is the official 10-minute demo INTRO: Party Land only, its own
	// two SHOWTEXT pages, one player started directly, and QUIT returning
	// to INTRO. ClosingText is the
	// launcher's exit text shown after Esc in the selector, when supplied.
	Demo        bool
	ClosingText []string
}

func New(store Store, factory Factory) (*Model, error) {
	var defaults [4]Scores
	for i := range defaults {
		defaults[i] = Defaults(i + 1)
	}
	return newWithDefaults(store, factory, defaults)
}

func newWithDefaults(store Store, factory Factory, defaults [4]Scores) (*Model, error) {
	m := &Model{Mode: Startup, Settings: settings.Defaults(), Selected: 1, Store: store, Factory: factory, Counter: 540, Rank: -1}
	for i := range m.Scores {
		var e error
		if store == nil {
			m.Scores[i] = defaults[i]
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

// restartIntro is a fresh INTRO process after the demo's QUIT.
func (m *Model) restartIntro() {
	m.End = Completed
	m.Session, m.sessionSynced = nil, false
	m.Mode, m.Counter = Startup, 540
	m.Segment, m.SegmentTick, m.IntroClock = 0, 0, 0
	m.Page, m.PreviousPage, m.TextPage, m.Reveal, m.SidebarTick = 0, 0, 0, 0, 0
}
func (m *Model) startPlayers(count int) error {
	s, e := m.tableFactory()(m.Scores[m.Selected-1][0].Digits)
	if e != nil {
		return e
	}
	if g, ok := s.(interface{ CarryLoadedTableState(any) }); ok {
		g.CarryLoadedTableState(m.Session)
	}
	// LATE_RASTER_INTERRUPT_DEMO starts FIRST_NO_OF_PLAYERSTS in the same
	// loaded VGA memory. NEW_BALL's VISAKEYS branch retains this program.
	if next, ok := s.(interface{ MatrixDisplay() *presentation.Display }); ok {
		if old, ok := m.Session.(interface{ MatrixDisplay() *presentation.Display }); ok {
			var names, scores [4]string
			for i, e := range m.Scores[m.Selected-1] {
				names[i], scores[i] = string(e.Name[:]), e.Digits.String()
			}
			matrix := old.MatrixDisplay().Attract(m.Counter, names, scores)
			if m.cheatTimeline != nil {
				matrix = m.cheatTimeline.Display()
			}
			if m.End == Completed && m.cheatTimeline == nil && m.gameOverTimeline == nil {
				players := []string{m.Final.String()}
				if len(m.scoreQueue) > 0 {
					players = make([]string, len(m.scoreQueue))
					for i, score := range m.scoreQueue {
						players[i] = score.String()
					}
				}
				matrix = old.MatrixDisplay().GameOverPlayers(m.Counter, players, names, scores)
			}
			if m.gameOverTimeline != nil && m.cheatTimeline == nil {
				matrix = m.gameOverTimeline.Display()
			}
			next.MatrixDisplay().CarryMemory(matrix)
		}
	}
	m.Session = s
	if g, ok := s.(interface{ SetCheatRules(bool, bool, int) }); ok {
		g.SetCheatRules(m.cheatTiltDisabled, m.cheatFastBall, m.cheatBalls)
	}
	m.cheatTimeline = nil
	m.gameOverTimeline = nil
	m.sourceScoreStage = 0
	if g, ok := s.(interface{ StartPlayers(int) }); ok {
		g.StartPlayers(count)
	}
	m.scoreQueue = nil
	m.highscorePlayed = false
	m.scorePlayer = 1
	m.selectionOpen, m.selectionDelay = !m.Demo, 15
	m.Mode = Playing
	m.End = Active
	return nil
}
func (m *Model) loadTable(table int) error {
	if e := m.saveSettings(); e != nil {
		return e
	}
	m.Selected = table
	m.cheatDecoder = gameplay.CheatDecoder{}
	m.cheatTimeline = nil
	m.cheatTiltDisabled, m.cheatFastBall, m.cheatBalls = false, false, 0
	m.Session = nil
	m.gameOverTimeline = nil
	m.sourceScoreStage = 0
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
		if m.sourceScoreStage == 1 {
			m.entrySetup = true
			m.sourceEntry(false)
		}
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

// ALFA_KEYS matches all four pinned DOS PRGs: A-Z and Space as '*'.
// Top-row numeric scans 2..11 are zero/rejected; see docs/dos-initials-audit.md.
func Initial(k Key) byte {
	const keys = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00QWERTYUIOP\x00\x00\x00\x00ASDFGHJKL\x00\x00\x00\x00\x00ZXCVBNM\x00\x00\x00\x00\x00\x00*"
	if int(k) < len(keys) {
		return keys[k]
	}
	return 0
}
func (m *Model) Update(in Input) error {
	m.Tick++
	m.sessionSynced = false
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
	// READ_KEYBOARD observes CLOSE1's ADDPLAYERS write before accepting F1..F8.
	if m.Mode == Playing {
		if g, ok := m.Session.(interface{ PlayerSelectionReady() bool }); ok && !g.PlayerSelectionReady() {
			m.selectionOpen = false
		}
	}
	initialMode := m.Mode
	if m.sourceScoreStage == 1 {
		if d, ok := m.Session.(interface{ MatrixDisplay() *presentation.Display }); ok {
			d.MatrixDisplay().Flash()
		}
		if m.Mode == Initials && m.entrySetup {
			// GET_IT_FROM_KEYBOARD inserts the row and clears SCAN_CODE.
			m.entrySetup = false
			return nil
		}
	}
	keys := in.Keys
	if m.sourceScoreStage == 1 && m.Mode == Initials && len(keys) > 1 {
		keys = keys[len(keys)-1:]
	} // SCAN_CODE is one byte.
	for _, k := range keys {
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
				if len(m.ClosingText) > 0 {
					m.Mode = Closing // the launcher prints its text, then waits for Esc
					return nil
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
			if m.Demo && k >= F2 && k <= F4 {
				break // the demo selector has no branches to the unavailable tables
			}
			if k >= F1 && k <= F4 {
				if e := m.loadTable(int(k-F1) + 1); e != nil {
					return e
				}
				if m.Demo && m.Mode == TableAttract {
					// Demo TABLE1 init stores F1 as the start key (mov [3813],3b
					// at 0x3a74): a one-player game starts with no attract wait.
					if e := m.startPlayers(1); e != nil {
						return e
					}
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
		case Closing:
			if k == Escape {
				m.Mode = Quit
				m.End = ProgramQuit
				return nil
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
			} else if program := m.cheatDecoder.Make(uint8(k)); program != "" {
				m.originalCheat(program)
			}
		case Playing:
			if matrixTestCheat(m, k) {
				continue
			}
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
			} else if k == Escape && m.SessionReady() && !m.Demo { // the demo omits the chute quit route
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
				if m.sourceScoreStage == 1 {
					m.sourceEntry(false)
				}
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
		m.sessionSynced = true
		if e := m.Session.Sync(in.controls()); e != nil {
			return e
		}
		matrixTestTrace(m)
		if g, ok := m.Session.(interface{ DemoFinished() bool }); ok && g.DemoFinished() {
			// Linked QUIT(0): TABLE1 exits and the launcher runs INTRO again.
			m.restartIntro()
			return nil
		}
		if d, done := m.Session.Result(); done {
			m.Final = d
			m.End = Completed
			m.Mode = GameEnd
			m.Counter = 5
			if src, ok := m.Session.(interface{ ScoreEntryPending() bool }); ok && src.ScoreEntryPending() {
				m.sourceScoreStage = 1
				m.loadScoreQueue()
				m.Counter = 1 // First SPINTSEL_IN_HIGH visit is next sync.
			}
		}
	case GameEnd:
		if m.sourceScoreStage == 2 {
			m.sessionSynced = true
			if e := m.Session.Sync(physics.Inputs{}); e != nil {
				return e
			}
			if _, done := m.Session.Result(); done {
				m.startGameOverTimeline()
			}
			break
		}
		m.Counter--
		if m.Counter == 0 {
			if m.sourceScoreStage == 0 {
				m.loadScoreQueue()
			}
			m.nextScore()
		}
	case TableAttract:
		m.Counter++
		if m.cheatTimeline != nil {
			m.cheatTimeline.Tick()
		}
		if m.gameOverTimeline != nil {
			m.gameOverTimeline.Tick()
		}
	case EntryWait:
		m.Counter--
		if m.sourceScoreStage == 1 && m.Counter == 30 {
			m.sourceEntry(true)
		}
		if m.Counter <= 2 {
			if m.sourceScoreStage == 1 {
				m.Mode = GameEnd
				m.Counter = 1
			} else {
				m.nextScore()
			}
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
	// The demo list has two entries: welcome and availability.
	if m.Demo {
		m.TextPage = m.TextPage%2 + 1
	} else {
		m.TextPage = m.TextPage%10 + 1
	}
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
	names := [...]string{"startup", "selector", "selector text", "options", "table attract", "playing", "paused", "quit question", "game over", "initials", "entry wait", "quit", "closing"}
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
		if m.sourceScoreStage == 1 {
			m.Mode = GameEnd
			m.Counter = 1
			return
		}
	}
	if m.sourceScoreStage == 1 {
		if !m.highscorePlayed {
			m.Session.Cue("S_GAMEOVER")
		}
		m.Session.(interface{ FinishScoreEntry() }).FinishScoreEntry()
		m.sourceScoreStage = 2
		m.Mode = GameEnd
		return
	}
	m.attract()
	if !m.highscorePlayed {
		m.Session.Cue("S_GAMEOVER")
	}
}

func (m *Model) loadScoreQueue() {
	if s, ok := m.Session.(interface{ PlayerScores() []tablelogic.Decimal }); ok {
		m.scoreQueue = s.PlayerScores()
	} else {
		m.scoreQueue = []tablelogic.Decimal{m.Final}
	}
	m.scorePlayer = 0
	m.highscorePlayed = false
}
func (m *Model) sourceEntry(stars bool) {
	if d, ok := m.Session.(interface{ MatrixDisplay() *presentation.Display }); ok {
		d.MatrixDisplay().ScoreEntry(m.scorePlayer, m.Entry, stars)
	}
}
func (m *Model) startGameOverTimeline() {
	var names, scores [4]string
	for i, e := range m.Scores[m.Selected-1] {
		names[i], scores[i] = string(e.Name[:]), e.Digits.String()
	}
	players := make([]string, len(m.scoreQueue))
	for i, d := range m.scoreQueue {
		players[i] = d.String()
	}
	if s, ok := m.Session.(interface {
		GameOverTimeline([]string, [4]string, [4]string) *presentation.Timeline
	}); ok {
		m.gameOverTimeline = s.GameOverTimeline(players, names, scores)
		m.gameOverTimeline.Tick() // Demo task precedes NODOT in this source sync.
	}
	m.sourceScoreStage = 0
	m.Mode = TableAttract
	m.Counter = 0
}
