package partyland

import (
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"
)

// PlayerState contains only values written by VARS_2_P_STRUC.
// The engine, jackpot, matrix, audio, physics and tasks remain shared.
type PlayerState struct {
	Score, Bonus, SkillTunnel, SkillCyclone Decimal
	Cyclones                                uint16
	Progress                                [17]bool
}

func (g *Game) SavePlayerState() PlayerState {
	s := PlayerState{Score: g.Score, Bonus: g.Bonus, SkillTunnel: g.SkillTunnel, SkillCyclone: g.SkillCyclone, Cyclones: g.Cyclones}
	for i, n := range []int{1, 2, 4, 5, 6, 8, 9, 41, 38, 34, 31, 28, 42, 43, 44, 45, 46} {
		s.Progress[i] = g.Lights[n]
	}
	return s
}
func (g *Game) LoadPlayerState(s PlayerState) {
	g.Score = s.Score
	g.Bonus = s.Bonus
	g.SkillTunnel = s.SkillTunnel
	g.SkillCyclone = s.SkillCyclone
	g.Cyclones = s.Cyclones
	for i, n := range []int{1, 2, 4, 5, 6, 8, 9, 41, 38, 34, 31, 28, 42, 43, 44, 45, 46} {
		g.Lights[n] = s.Progress[i]
	}
}
func (g *Game) StartPlayers(count int) {
	g.Session.Initialize(count, g.SavePlayerState())
	g.playerText()
	// LATE_RASTER_INTERRUPT_DEMO dispatches this for every F1..F8 count.
	g.beginMatrix("FIRST_NO_OF_PLAYERSTS")
	g.sound("S_ADDPLAYER2")
}
func (g *Game) SelectPlayers(count int) {
	g.Session.Select(count)
	g.playerText()
	g.beginMatrix("NO_OF_PLAYERSTS")
	g.sound("S_ADDPLAYER2")
}
func (g *Game) PlayerScores() []tablelogic.Decimal {
	return g.Session.Scores(func(s PlayerState) tablelogic.Decimal { return s.Score }, g.Score)
}
func (g *Game) PlayerCount() int   { return g.Session.PlayerCount }
func (g *Game) CurrentPlayer() int { return g.Session.CurrentPlayer }
func (g *Game) playerText() {
	var balls []byte
	if g.timed != nil {
		balls = g.Display.Content.Texts["BALLSTEXT"]
	}
	g.Display.SetPlayers(g.Session.CurrentPlayer, g.Session.PlayerCount, int(g.BallNumber))
	g.demoPlayerText(balls)
}
func (g *Game) savePlayer() { g.Session.Save(g.SavePlayerState()) }
func (g *Game) advancePlayer() bool {
	g.savePlayer()
	if !g.Session.Advance(&g.BallNumber, g.totalBalls) {
		return false
	}
	if g.Session.PlayerCount > 1 {
		g.LoadPlayerState(g.Session.Load())
	}
	g.playerText()
	return true
}
func (g *Game) selectMatch(digit uint8) bool {
	g.savePlayer()
	if !g.Session.StartMatch(digit, func(s PlayerState) tablelogic.Decimal { return s.Score }) {
		return false
	}
	g.LoadPlayerState(g.Session.Load())
	g.playerText()
	g.matchBall = true
	return true
}
func (g *Game) nextMatch() bool {
	g.savePlayer()
	if !g.Session.NextMatch() {
		return false
	}
	g.LoadPlayerState(g.Session.Load())
	g.playerText()
	return true
}

func (g *Game) matchDigits() []byte {
	scores := g.PlayerScores()
	digits := make([]byte, len(scores))
	for i, s := range scores {
		digits[i] = s[10]
	}
	return digits
}
func (g *Game) anyMatch(digit uint8) bool {
	for _, d := range g.matchDigits() {
		if d == digit {
			return true
		}
	}
	return false
}
func (g *Game) matchStart() {
	if g.Session.PlayerCount == 1 {
		g.Display.MatchStart(g.Score[10])
		return
	}
	g.Display.MatchPlayers(g.matchDigits(), nil)
}
func (g *Game) matchWin(digit uint8) {
	if g.Session.PlayerCount == 1 {
		g.Display.MatchWin(g.Score[10])
		return
	}
	g.Display.MatchPlayers(g.matchDigits(), &digit)
}

func (g *Game) PlayerSelectionReady() bool { return g.Session.SelectionOpen && g.inChute }

func (g *Game) MatrixDisplay() *presentation.Display { return g.Display }
