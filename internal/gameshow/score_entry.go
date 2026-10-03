package gameshow

import "pinballfantasies/internal/presentation"

// ScoreEntryPending marks the source SPINTSEL_IN_HIGH handoff, rather than
// a host-generated game-over pause. Storage and keyboard input live in frontend.
func (g *Game) ScoreEntryPending() bool { return g.matrix.op == "_CHECK_HIGH" }
func (g *Game) FinishScoreEntry() {
	g.Phase = BallLost
	g.matrixDispatch() // SI=0 dispatches _MATRIXLGT in the same source sync.
}

func (g *Game) GameOverTimeline(players []string, names, scores [4]string) *presentation.Timeline {
	return g.Display.ContinueGameOver(players, names, scores)
}

// TO_DEMO_FROM_GAME is a task, visited before matrix dispatch on the next sync.
func (g *Game) enterDemo() {
	g.Phase = GameOver
	g.ScreenForce = g.Physics.BottomRaster()
	for n := 1; n < len(g.Lights); n++ {
		g.light(n, false)
	}
	g.Score = Decimal{} // ZEROSCORE; PLAYER_AREA retains completed scores.
}
