package frontend

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"testing"
)

type memoryStore struct {
	entries [4]Scores
	writes  int
}

func (s *memoryStore) Load(t int) (Scores, error) { return s.entries[t-1], nil }
func (s *memoryStore) Save(t int, v Scores) error { s.entries[t-1] = v; s.writes++; return nil }

type fakeSession struct {
	tick, releases int
	done, chute    bool
	score          partyland.Decimal
	cue            string
}

func (s *fakeSession) Sync(in physics.Inputs) error {
	s.tick++
	if in.Release {
		s.releases++
	}
	return nil
}
func (s *fakeSession) Release(uint8, uint8)              { s.releases++ }
func (s *fakeSession) Frame() *image.RGBA                { return image.NewRGBA(image.Rect(0, 0, 320, 383)) }
func (s *fakeSession) Result() (partyland.Decimal, bool) { return s.score, s.done }
func (s *fakeSession) PCM() []byte                       { return nil }
func (s *fakeSession) Cue(v string)                      { s.cue = v }
func (s *fakeSession) InChute() bool                     { return s.chute }
func setup(t *testing.T) (*Model, *memoryStore, *int) {
	t.Helper()
	store := &memoryStore{}
	for i := range store.entries {
		store.entries[i] = Defaults(i + 1)
	}
	calls := new(int)
	m, e := New(store, func(top partyland.Decimal) (Session, error) {
		if top != store.entries[0][0].Digits {
			t.Fatal("top score not supplied")
		}
		*calls++
		return &fakeSession{chute: true}, nil
	})
	if e != nil {
		t.Fatal(e)
	}
	return m, store, calls
}
func update(t *testing.T, m *Model, in Input) {
	t.Helper()
	if e := m.Update(in); e != nil {
		t.Fatal(e)
	}
}
func key(t *testing.T, m *Model, k Key) { t.Helper(); update(t, m, Input{Keys: []Key{k}}) }
func enterGame(t *testing.T, m *Model) {
	t.Helper()
	key(t, m, Space)
	key(t, m, F1)
	if m.Mode != TableAttract {
		t.Fatal("table load should enter attract")
	}
	key(t, m, Enter)
	if m.Mode != Playing {
		t.Fatal("single player start")
	}
}
func TestStartupAndSelector(t *testing.T) {
	m, _, calls := setup(t)
	if m.Mode != Startup || m.Selected != 1 || m.Hz() != 60 {
		t.Fatal("startup")
	}
	for i := 0; i < 2000 && m.Mode == Startup; i++ {
		update(t, m, Input{})
	}
	if m.Mode != Selector || *calls != 0 || m.IntroClock != 1822 {
		t.Fatalf("startup progression: %v clock=%d", m.Mode, m.IntroClock)
	}
	for i := 0; i < 580; i++ {
		update(t, m, Input{})
	}
	if m.Mode != SelectorText || m.Page != 1 || m.TextPage != 1 {
		t.Fatal("selector timer")
	}
	key(t, m, Space)
	if m.Mode != SelectorText || m.Counter != 420 {
		t.Fatal("INTRO does not read skip keys during text reveal/fade")
	}
	for m.TextTick < 45 {
		update(t, m, Input{})
	}
	key(t, m, Space)
	if m.Mode != SelectorText || m.Counter != 0 || m.TextExitTick != 0 {
		t.Fatal("skip must begin the source exit fade, not jump to previews")
	}
	for i := 0; i < 20; i++ {
		update(t, m, Input{})
	}
	if m.Mode != SelectorText || m.TextExitTick != 20 {
		t.Fatal("fade3b must consume twenty syncs")
	}
	update(t, m, Input{})
	if m.Mode != Selector || m.Page != 1 {
		t.Fatal("return after source exit fade")
	}
	key(t, m, Enter)
	if m.Page != 0 || m.Mode != Selector {
		t.Fatal("enter changes preview pair")
	}
}
func TestTableScopeAndHandoff(t *testing.T) {
	m, _, calls := setup(t)
	key(t, m, Space)
	for _, k := range []Key{F2, F3, F4} {
		key(t, m, k)
		if m.Mode != Selector || m.Session != nil || *calls != 0 || m.Selected != int(k-F1)+1 {
			t.Fatal("unavailable table launched")
		}
	}
	key(t, m, F1)
	if m.Mode != TableAttract || *calls != 1 || m.Selected != 1 {
		t.Fatal("Party Land load")
	}
	key(t, m, F1)
	s := m.Session.(*fakeSession)
	if *calls != 2 || m.Mode != Playing || s.tick != 0 {
		t.Fatal("start leaks input")
	}
	update(t, m, Input{Release: true, Left: true})
	if s.tick != 1 || s.releases != 1 {
		t.Fatal("game handoff")
	}
}
func TestPauseResumeAbortAndClose(t *testing.T) {
	m, _, _ := setup(t)
	enterGame(t, m)
	s := m.Session.(*fakeSession)
	update(t, m, Input{})
	key(t, m, P)
	for i := 0; i < 200; i++ {
		update(t, m, Input{Release: true, Left: true})
	}
	if s.tick != 1 || s.releases != 0 || m.Mode != Paused {
		t.Fatal("pause advanced session")
	}
	key(t, m, Key(127))
	update(t, m, Input{})
	if s.tick != 2 || m.Mode != Playing {
		t.Fatal("resume")
	}
	for i := 0; i < 30; i++ {
		update(t, m, Input{})
	}
	key(t, m, P)
	key(t, m, Escape)
	if m.Mode != QuitQuestion || s.tick != 32 {
		t.Fatal("pause quit prompt")
	}
	key(t, m, Key(49))
	if m.Mode != Playing {
		t.Fatal("N resumes")
	}
	s.chute = false
	key(t, m, Escape)
	if m.Mode != Playing {
		t.Fatal("ingame Escape outside chute must be ignored")
	}
	s.chute = true
	key(t, m, Escape)
	if m.Mode != TableAttract || m.End != Aborted {
		t.Fatal("chute abort")
	}
	key(t, m, Escape)
	key(t, m, Key(21))
	if m.Mode != Selector || m.Session != nil {
		t.Fatal("Y exits table")
	}
	update(t, m, Input{Close: true})
	if m.Mode != Quit || m.End != ProgramQuit {
		t.Fatal("window close")
	}
}
func TestPausedQuitReturnsSelectorWithoutScore(t *testing.T) {
	m, store, _ := setup(t)
	enterGame(t, m)
	key(t, m, P)
	key(t, m, Escape)
	key(t, m, Key(21))
	if m.Mode != Selector || m.End != Aborted || store.writes != 0 {
		t.Fatal("aborted game inserted high score")
	}
}
func TestScoresAndEntry(t *testing.T) {
	for _, tc := range []struct {
		score uint64
		rank  int
	}{{5_000_000, -1}, {5_000_001, 3}, {10_000_000, 3}, {25_000_000, 2}, {50_000_000, 1}, {50_000_001, 0}} {
		t.Run(partyland.Number(tc.score).String(), func(t *testing.T) {
			m, store, _ := setup(t)
			enterGame(t, m)
			s := m.Session.(*fakeSession)
			s.score = partyland.Number(tc.score)
			s.done = true
			update(t, m, Input{})
			if m.Mode != GameEnd || m.Final != s.score {
				t.Fatal("game over result")
			}
			for i := 0; i < 5; i++ {
				update(t, m, Input{})
			}
			if m.Rank != tc.rank {
				t.Fatalf("rank=%d", m.Rank)
			}
			if tc.rank < 0 {
				if m.Mode != TableAttract || s.cue != "S_GAMEOVER" || store.writes != 0 {
					t.Fatal("nonqualification")
				}
				return
			}
			if m.Mode != Initials || s.cue != "S_GAMEOVER2" {
				t.Fatal("qualification")
			}
			for _, k := range []Key{Escape, Enter, 14, 2} {
				key(t, m, k)
			}
			if m.Entered != 0 {
				t.Fatal("invalid characters accepted")
			}
			for _, k := range []Key{30, 48, 46} {
				key(t, m, k)
			}
			if m.Mode != EntryWait || store.writes != 1 || m.Scores[0][tc.rank].Name != [3]byte{'A', 'B', 'C'} {
				t.Fatal("three initials")
			}
			for i := 0; i < 58; i++ {
				update(t, m, Input{})
			}
			if m.Mode != TableAttract || m.End != Completed {
				t.Fatal("return to attract")
			}
			key(t, m, Escape)
			key(t, m, 21)
			if m.Mode != Selector {
				t.Fatal("return to selector")
			}
		})
	}
}
func TestBinaryScores(t *testing.T) {
	s := Defaults(1)
	s.Insert(1, partyland.Number(40_000_000), [3]byte{'A', 'B', 'C'})
	if s[2].Digits.Uint64() != 25_000_000 || s[3].Digits.Uint64() != 10_000_000 {
		t.Fatal("insertion shift")
	}
	b := s.MarshalBinary()
	decoded, e := DecodeScores(b)
	if e != nil || decoded != s {
		t.Fatal("round trip", e)
	}
	for _, bad := range [][]byte{b[:63], append([]byte{10}, b[1:]...), append(append([]byte{}, b[:15]...), append([]byte{1}, b[16:]...)...)} {
		if _, e := DecodeScores(bad); e == nil {
			t.Fatal("invalid record accepted")
		}
	}
}
func TestPersistenceSeedsWithoutChangingAssets(t *testing.T) {
	dir, seed := t.TempDir(), t.TempDir()
	original := Defaults(1).MarshalBinary()
	if e := os.WriteFile(filepath.Join(seed, "TABLE1.HI"), original, 0600); e != nil {
		t.Fatal(e)
	}
	store := FileStore{dir, seed}
	s, e := store.Load(1)
	if e != nil {
		t.Fatal(e)
	}
	s.Insert(0, partyland.Number(99_000_000), [3]byte{'A', 'B', 'C'})
	if e := store.Save(1, s); e != nil {
		t.Fatal(e)
	}
	got, e := store.Load(1)
	if e != nil || got != s {
		t.Fatal("persistent scores", e)
	}
	b, _ := os.ReadFile(filepath.Join(seed, "TABLE1.HI"))
	if !bytes.Equal(b, original) {
		t.Fatal("installation seed overwritten")
	}
	files, _ := os.ReadDir(dir)
	if len(files) != 1 {
		t.Fatal("temporary file left")
	}
	fallback, e := store.Load(2)
	if e != nil || fallback != Defaults(2) {
		t.Fatal("missing file defaults")
	}
}
func TestKeyTableAgainstOriginal(t *testing.T) {
	for table, offset := range []int{0x1d396, 0x1c5cb, 0x1ba70, 0x1a2f4} {
		path := fmt.Sprintf("../../TABLE%d.PRG", table+1)
		t.Run(fmt.Sprint(table+1), func(t *testing.T) {
			testinputs.Require(t, path)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for scan := 0; scan < 128; scan++ {
				if got := Initial(Key(scan)); got != data[offset+scan] {
					t.Fatalf("scan %d %q != %q", scan, got, data[offset+scan])
				}
			}
		})
	}
}

// All four pinned DOS PRGs reject the normal top-row numeric makes.
func TestInitialsRejectDOSDigits(t *testing.T) {
	m, store, _ := setup(t)
	m.Mode = Initials
	for scan := Key(2); scan <= 11; scan++ {
		key(t, m, scan)
		if Initial(scan) != 0 || m.Entered != 0 || store.writes != 0 || m.Mode != Initials {
			t.Fatalf("numeric scan %d changed initials", scan)
		}
	}
	for _, sequence := range []struct {
		keys []Key
		want [3]byte
	}{
		{[]Key{30, 2, 48}, [3]byte{'A', 'B', 0}},
		{[]Key{8, 22, 25}, [3]byte{'U', 'P', 0}},
		{[]Key{19, 3, 32}, [3]byte{'R', 'D', 0}},
	} {
		m, store, _ := setup(t)
		m.Mode = Initials
		for _, scan := range sequence.keys {
			key(t, m, scan)
		}
		if m.Entry != sequence.want || m.Entered != 2 || m.Mode != Initials || store.writes != 0 {
			t.Fatalf("mixed input: entry=%q entered=%d mode=%v writes=%d", m.Entry, m.Entered, m.Mode, store.writes)
		}
	}
}

func TestNativeSessionSuspension(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	for _, k := range []Key{Space, F1, F1} {
		if e := r.Update(Input{Keys: []Key{k}}); e != nil {
			t.Fatal(e)
		}
	}
	g := r.Model.Session.(*partyland.Game)
	for i := 0; i < 120; i++ {
		if e := r.Update(Input{Release: i == 1, Left: i%10 < 3}); e != nil {
			t.Fatal(e)
		}
	}
	if e := r.Update(Input{Keys: []Key{P}}); e != nil {
		t.Fatal(e)
	}
	tick, syncs, ball, score, frames := g.Tick, g.Physics.Syncs, g.Physics.Ball, g.Score, g.Playback.Frames
	before := r.Frame().Pix
	for i := 0; i < 100; i++ {
		if e := r.Update(Input{Left: true, Right: true, Release: true}); e != nil {
			t.Fatal(e)
		}
		if len(r.PCM) != 0 {
			t.Fatal("audio advanced while paused")
		}
		if !bytes.Equal(before, r.Frame().Pix) {
			t.Fatal("pause render changes")
		}
	}
	if g.Tick != tick || g.Physics.Syncs != syncs || g.Physics.Ball != ball || g.Score != score || g.Playback.Frames != frames {
		t.Fatal("native gameplay mutated during pause")
	}
	if e := r.Update(Input{Keys: []Key{P}}); e != nil {
		t.Fatal(e)
	}
	if e := r.Update(Input{}); e != nil {
		t.Fatal(e)
	}
	if g.Tick != tick+1 {
		t.Fatal("native resume")
	}
}
func TestNativeGameOverAndRestart(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	for _, k := range []Key{Space, F1, Enter} {
		if e := r.Update(Input{Keys: []Key{k}}); e != nil {
			t.Fatal(e)
		}
	}
	g := r.Model.Session.(*partyland.Game)
	g.Score = partyland.Number(6_000_000)
	g.ScoreChanged = true
	// Invoke PF3's existing drain callback, then let PF4's bonus/new-ball/match
	// scheduler decide when the real three-ball game is complete.
	drained := false
	for i := 0; i < 16000 && r.Model.Mode == Playing; i++ {
		if g.Phase == partyland.Playing && !drained {
			g.ScoreChanged = true
			g.Physics.OnEvent(physics.Event{Kind: physics.EventDrain})
			drained = true
		}
		if g.Phase == partyland.NewBall {
			drained = false
		}
		if e := r.Update(Input{}); e != nil {
			t.Fatal(e)
		}
	}
	if r.Model.Mode != GameEnd {
		t.Fatalf("native did not hand off: mode %v ball %d phase %v audio %+v events %+v", r.Model.Mode, g.BallNumber, g.Phase, g.Audio, g.Events)
	}
	for i := 0; i < 5; i++ {
		if e := r.Update(Input{}); e != nil {
			t.Fatal(e)
		}
	}
	if r.Model.Mode != Initials || r.Model.Final != g.Score {
		t.Fatal("native final score")
	}
	for _, k := range []Key{30, 48, 46} {
		if e := r.Update(Input{Keys: []Key{k}}); e != nil {
			t.Fatal(e)
		}
	}
	for i := 0; i < 62; i++ {
		if e := r.Update(Input{}); e != nil {
			t.Fatal(e)
		}
	}
	if e := r.Update(Input{Keys: []Key{F1}}); e != nil {
		t.Fatal(e)
	}
	newGame := r.Model.Session.(*partyland.Game)
	if reflect.DeepEqual(newGame, g) || newGame.Score != (partyland.Decimal{}) || newGame.BallNumber != 1 || newGame.Tick != 0 {
		t.Fatal("new session inherited finished state")
	}
}

func TestPF6SessionKeepsGameplayOracle(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, e := Load("../..", nil)
	if e != nil {
		t.Fatal(e)
	}
	for _, k := range []Key{Space, F1, F1} {
		if e := r.Update(Input{Keys: []Key{k}}); e != nil {
			t.Fatal(e)
		}
	}
	g := r.Model.Session.(*partyland.Game)
	for i := 0; i < 1200; i++ {
		if i == 100 || i == 700 {
			r.Model.Session.Release(32, 0)
		}
		if e := r.Update(Input{Left: i%93 < 18, Right: i%71 < 14}); e != nil {
			t.Fatal(e)
		}
	}
	if g.Score.Uint64() != 2_300_000 || g.BallNumber != 2 || g.Tick != 1200 {
		t.Fatalf("PF6 gameplay oracle score=%s ball=%d ticks=%d", g.Score, g.BallNumber, g.Tick)
	}
}
func TestCloseEveryLifecycleMode(t *testing.T) {
	for mode := Startup; mode <= EntryWait; mode++ {
		m, store, _ := setup(t)
		m.Mode = mode
		update(t, m, Input{Close: true})
		if m.Mode != Quit || m.End != ProgramQuit || store.writes != 0 {
			t.Fatalf("close in %s", mode)
		}
	}
}
func TestEntryUsesSourceSpaceStar(t *testing.T) {
	m, store, _ := setup(t)
	enterGame(t, m)
	m.Session.(*fakeSession).done = true
	m.Session.(*fakeSession).score = partyland.Number(60_000_000)
	update(t, m, Input{})
	for i := 0; i < 5; i++ {
		update(t, m, Input{})
	}
	update(t, m, Input{Keys: []Key{Space, Space, Space}})
	if m.Mode != EntryWait || m.Entry != [3]byte{'*', '*', '*'} || store.writes != 1 {
		t.Fatal("Space initials do not follow ALFA_KEYS")
	}
}

func TestReloadTableStartsFreshAttract(t *testing.T) {
	m, store, _ := setup(t)
	m.End = Completed
	m.Final = partyland.Number(123456)
	if e := m.loadTable(1); e != nil {
		t.Fatal(e)
	}
	if m.Mode != TableAttract || m.End != Active || m.Final != (partyland.Decimal{}) || m.Counter != 0 {
		t.Fatal("fresh table replayed previous GameOver program")
	}
	if store.writes != 0 {
		t.Fatal("loading attract changed high scores")
	}
}

func TestFocusLossPausesNativeGameplay(t *testing.T) {
	testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, err := Load("../..", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []Key{Space, F1, F1} {
		if err := r.Update(Input{Keys: []Key{key}}); err != nil {
			t.Fatal(err)
		}
	}
	g := r.Model.Session.(*partyland.Game)
	r.Model.PauseDelay = 30 // focus loss must not obey manual pause debounce
	tick, syncs, ball, score, frames := g.Tick, g.Physics.Syncs, g.Physics.Ball, g.Score, g.Playback.Frames
	for i := 0; i < 3; i++ {
		if err := r.Update(Input{FocusLost: true, Keys: []Key{P, Enter}, Left: true, Right: true, Release: true}); err != nil {
			t.Fatal(err)
		}
		if r.Model.Mode != Paused || len(r.PCM) != 0 {
			t.Fatal("focus loss failed to pause, or queued keys resumed it")
		}
	}
	// Focus return has no make event and must leave the pause in place.
	if err := r.Update(Input{}); err != nil {
		t.Fatal(err)
	}
	if r.Model.Mode != Paused || g.Tick != tick || g.Physics.Syncs != syncs || g.Physics.Ball != ball || g.Score != score || g.Playback.Frames != frames {
		t.Fatal("gameplay/audio advanced during focus pause")
	}
	if err := r.Update(Input{Keys: []Key{P}}); err != nil {
		t.Fatal(err)
	}
	if r.Model.Mode != Playing {
		t.Fatal("ordinary pause resume failed")
	}
	if err := r.Update(Input{}); err != nil {
		t.Fatal(err)
	}
	if g.Tick != tick+1 {
		t.Fatal("gameplay failed to resume")
	}
	// Close remains authoritative when it arrives with focus loss.
	if err := r.Update(Input{FocusLost: true, Close: true}); err != nil {
		t.Fatal(err)
	}
	if r.Model.Mode != Quit {
		t.Fatal("focus pause swallowed close")
	}
}
