package platform

import (
	"fmt"
	"os"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"pinballfantasies/internal/tablelogic"
	"strconv"
	"testing"
	"time"
)

// TestHotseatLiveHost uses real host keyboard/focus/fullscreen events. The
// external driver launches each physical ball; this test places a launched ball
// below the drain and seeds a distinct score to bound the journey. Neither
// fixture is available in the production executable.
func TestHotseatLiveHost(t *testing.T) {
	if os.Getenv("PF121_LIVE") != "1" {
		t.Skip("opt-in native host journey")
	}
	table, _ := strconv.Atoi(os.Getenv("PF121_TABLE"))
	r, err := frontend.LoadConfigured(os.Getenv("PF121_DATA"), nil, &settings.Store{Directory: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	r.Model.Settings.Music = 0
	if table == 4 {
		r.Model.Settings.ScrollMode = settings.ScrollOff
	} // Both hosts exercise full-table OFF as well as scrolling.
	a, err := OpenAudio()
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	lastTurn := ""
	turns := 0
	launchedAt := uint64(0)
	fixture := false
	nonzero := false
	paused := false
	fullscreen := false
	err = showFrontend(r, 90*time.Second, a, func(h *hostWindow) {
		if r.Model.Mode == frontend.Paused {
			paused = true
		}
		if h.presentation.IsFullscreen() {
			fullscreen = true
		}
		for _, b := range r.PCM {
			if b != 0 {
				nonzero = true
				break
			}
		}
		if r.Model.Mode != frontend.Playing {
			return
		}
		var p *physics.Game
		var score *tablelogic.Decimal
		var ball uint8
		var count, player int
		switch g := r.Model.Session.(type) {
		case *partyland.Game:
			p, score, ball, count, player = g.Physics, &g.Score, g.BallNumber, g.PlayerCount(), g.CurrentPlayer()
		case *speeddevils.Game:
			p, score, ball, count, player = g.Physics, &g.Score, g.BallNumber, g.PlayerCount(), g.CurrentPlayer()
		case *gameshow.Game:
			p, score, ball, count, player = g.Physics, &g.Score, g.BallNumber, g.PlayerCount(), g.CurrentPlayer()
		case *stones.Game:
			p, score, ball, count, player = g.Physics, &g.Score, g.BallNumber, g.PlayerCount(), g.CurrentPlayer()
		}
		if p == nil || count != 2 || r.Model.Selected != table {
			return
		}
		turn := fmt.Sprintf("P%d B%d", player, ball)
		if turn != lastTurn {
			lastTurn = turn
			turns++
			fixture = false
			launchedAt = 0
			fmt.Printf("PF121 TURN %s score=%s\n", turn, score.String())
			if turns == 2 && score.Uint64() != 0 {
				t.Error("P2 inherited P1 score")
			}
			if turns == 3 && score.Uint64() == 0 {
				t.Error("P1 score lost on wrap")
			}
			if turns == 5 {
				r.Update(frontend.Input{Close: true})
				return
			}
		}
		if p.Ball.VY < -1000 && launchedAt == 0 {
			launchedAt = r.Model.Tick
			fmt.Printf("PF121 LAUNCH %s\n", turn)
		}
		if launchedAt > 0 && r.Model.Tick > launchedAt+120 && !fixture {
			*score = tablelogic.Number(uint64(player)*10000 + uint64(ball)*10)
			if g, ok := r.Model.Session.(*partyland.Game); ok {
				g.ScoreChanged = true
			}
			p.SetBall(150, 620, 0, 1000, false)
			p.Ball.Hold = false
			fixture = true
			fmt.Printf("PF121 FIXTURE DRAIN %s score=%s\n", turn, score.String())
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if turns < 5 || !paused || !fullscreen || !nonzero {
		t.Fatalf("incomplete live journey turns=%d pause=%t fullscreen=%t PCM=%t", turns, paused, fullscreen, nonzero)
	}
	t.Logf("table=%d two rounds, independent scores, pause/focus, fullscreen, nonzero PCM; bounded drain/score fixtures; starts=%d resets=%d", table, a.Starts, a.Dropped)
}
