package frontend

import (
	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"pinballfantasies/internal/tablelogic"
	"pinballfantasies/internal/testinputs"
	"testing"
)

type hotseatSession interface {
	Session
	PlayerCount() int
	CurrentPlayer() int
	PlayerScores() []tablelogic.Decimal
}

func hotseatRuntime(t *testing.T) *Runtime {
	t.Helper()
	testinputs.Require(t, "../../INTRO.PRG", "../../TABLE1.PRG", "../../TABLE2.PRG", "../../TABLE3.PRG", "../../TABLE4.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.MOD", "../../TABLE2.MOD", "../../TABLE3.MOD", "../../TABLE4.MOD", "../../PINBALL.CFG")
	r, err := Load("../..", nil)
	if err != nil {
		t.Fatal(err)
	}
	key(t, r.Model, Space)
	return r
}
func TestHotseatSelectionAllTables(t *testing.T) {
	for table := 1; table <= 4; table++ {
		r := hotseatRuntime(t)
		m := r.Model
		key(t, m, Key(int(F1)+table-1))
		for _, count := range []int{1, 2, 3, 5, 6, 7, 8} {
			m.attract()
			key(t, m, Key(int(F1)+count-1))
			g := m.Session.(hotseatSession)
			if m.Mode != Playing || g.PlayerCount() != count || g.CurrentPlayer() != 1 {
				t.Fatalf("table %d players %d", table, count)
			}
		}
		m.attract()
		key(t, m, Enter)
		if m.Mode != TableAttract {
			t.Fatal("source Enter guard after eight-player game")
		}
		key(t, m, F1)
		m.attract()
		key(t, m, Enter)
		g := m.Session.(hotseatSession)
		if g.PlayerCount() != 1 {
			t.Fatal("Enter fresh game must start one player")
		}
		key(t, m, F8) // blocked for source ALLOW_ADDPL interval
		if g.PlayerCount() != 1 {
			t.Fatal("selection debounce")
		}
		for i := 0; i < 16; i++ {
			update(t, m, Input{})
		}
		key(t, m, Enter)
		if g.PlayerCount() != 2 {
			t.Fatal("Enter adds a player")
		}
		for i := 0; i < 16; i++ {
			update(t, m, Input{})
		}
		key(t, m, F5)
		if m.Mode != Playing || g.PlayerCount() != 5 {
			t.Fatal("F5 must select five players in table")
		}
		// First launch CLOSE1 ends selection for the whole game, even later chutes.
		m.selectionOpen = false
		for i := 0; i < 16; i++ {
			update(t, m, Input{})
		}
		key(t, m, F8)
		if g.PlayerCount() != 5 {
			t.Fatal("active game accepted player count")
		}
		tick := m.Tick
		key(t, m, P)
		if m.Mode != Paused {
			t.Fatal("pause")
		}
		key(t, m, F8)
		if m.Mode != Playing || g.PlayerCount() != 5 || m.Tick != tick+2 {
			t.Fatal("resume leaked selection key")
		}
		update(t, m, Input{FocusLost: true})
		if m.Mode != Paused {
			t.Fatal("focus loss")
		}
		key(t, m, P)
		m.attract()
		key(t, m, F1)
		if m.Session.(hotseatSession).PlayerCount() != 1 {
			t.Fatal("old count survived new game")
		}
	}
}
func TestHotseatHighScoresPlayerOrder(t *testing.T) {
	r := hotseatRuntime(t)
	m := r.Model
	key(t, m, F1)
	key(t, m, F3)
	g := m.Session.(*partyland.Game)
	g.Session.Players[0].Score = tablelogic.Number(60000000)
	g.Session.Players[1].Score = tablelogic.Number(80000000)
	g.Score = tablelogic.Number(60000000)
	g.Session.CurrentPlayer = 3
	m.Mode = GameEnd
	m.Counter = 1
	update(t, m, Input{})
	for player := 1; player <= 3; player++ {
		if m.Mode != Initials || m.scorePlayer != player {
			t.Fatalf("entry order P%d mode %v entry P%d", player, m.Mode, m.scorePlayer)
		}
		for _, k := range []Key{30, 48, 46} {
			key(t, m, k)
		}
		for i := 0; i < 60; i++ {
			update(t, m, Input{})
		}
	}
	if m.Mode != TableAttract {
		t.Fatal("score queue did not complete")
	}
	got := m.Scores[0]
	if got[0].Digits.Uint64() != 80000000 || got[1].Digits.Uint64() != 60000000 || got[2].Digits.Uint64() != 60000000 {
		t.Fatal("multiple entries/tie handling", got)
	}
	r.Frame()
}

// Cover frontend game-over gating using actual table sessions, not fake scores.
func TestHotseatGameOverAfterFinalPlayer(t *testing.T) {
	for table := 1; table <= 4; table++ {
		r := hotseatRuntime(t)
		m := r.Model
		key(t, m, Key(int(F1)+table-1))
		key(t, m, F2)
		switch g := m.Session.(type) {
		case *partyland.Game:
			g.Phase = partyland.GameOver
		case *speeddevils.Game:
			g.Phase = speeddevils.GameOver
		case *gameshow.Game:
			g.Phase = gameshow.GameOver
		case *stones.Game:
			g.Phase = stones.GameOver
		}
		update(t, m, Input{})
		if m.Mode != GameEnd {
			t.Fatal("frontend missed game over", table)
		}
		for i := 0; i < 5; i++ {
			update(t, m, Input{})
		}
		if m.Mode != TableAttract || len(m.scoreQueue) != 2 {
			t.Fatal("player scores lost", table)
		}
		key(t, m, F1)
		if m.Session.(hotseatSession).PlayerCount() != 1 {
			t.Fatal("multiplayer reset")
		}
	}
}

type hotseatCueSession struct {
	fakeSession
	cues []string
}

func (s *hotseatCueSession) Cue(v string) { s.cues = append(s.cues, v) }
func TestHotseatHighScoreJingleOnce(t *testing.T) {
	for _, qualify := range []bool{false, true} {
		s := &hotseatCueSession{}
		m := &Model{Session: s, Selected: 1, Scores: [4]Scores{Defaults(1)}}
		if qualify {
			m.scoreQueue = []tablelogic.Decimal{tablelogic.Number(60000000), tablelogic.Number(80000000)}
		} else {
			m.scoreQueue = []tablelogic.Decimal{{}, {}}
		}
		m.nextScore()
		for m.Mode == Initials {
			m.Scores[0].Insert(m.Rank, m.Final, [3]byte{'A', 'B', 'C'})
			m.nextScore()
		}
		want := "S_GAMEOVER"
		if qualify {
			want = "S_GAMEOVER2"
		}
		if len(s.cues) != 1 || s.cues[0] != want {
			t.Fatalf("qualify=%t cues=%v", qualify, s.cues)
		}
	}
}

func TestSourceStartupRetainsAttractMemoryForEveryPlayerCount(t *testing.T) {
	// FANTASIE LATE_RASTER_INTERRUPT_DEMO calls DO_MATRIX
	// FIRST_NO_OF_PLAYERSTS for all F1..F8. NEW_BALL's VISAKEYS branch
	// does not draw SHOWPLAYERSTS or clear VGA dots at that dispatch.
	for table := 1; table <= 4; table++ {
		for count := 1; count <= 8; count++ {
			r := hotseatRuntime(t)
			m := r.Model
			key(t, m, Key(int(F1)+table-1))
			m.Counter = 317
			old := m.Session.(interface{ MatrixDisplay() *presentation.Display }).MatrixDisplay()
			var names, scores [4]string
			for i, e := range m.Scores[table-1] {
				names[i], scores[i] = string(e.Name[:]), e.Digits.String()
			}
			want := old.Attract(m.Counter, names, scores)
			key(t, m, Key(int(F1)+count-1))
			got := m.Session.(interface{ MatrixDisplay() *presentation.Display }).MatrixDisplay()
			if got.Dots != want.Dots || !got.On {
				t.Fatalf("table%d players%d startup discarded VGA memory", table, count)
			}
			if got.Argument(0) != got.Content.Commands[got.Content.Labels["FIRST_NO_OF_PLAYERSTS"]].Arg(0) {
				t.Fatal("wrong initial matrix program")
			}
			// The original source's selection program must eventually reach
			// WAIT_GAME_ON even for one player; it is not an idle score panel.
			for tick := 0; tick < 20; tick++ {
				update(t, m, Input{})
			}
			if got.CurrentOperation() != "_WAIT_GAME_ON" {
				t.Fatalf("table%d players%d source selection not held", table, count)
			}
		}
	}
}
