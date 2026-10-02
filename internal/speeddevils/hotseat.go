package speeddevils

import "pinballfantasies/internal/tablelogic"

// PlayerState contains only values written by VARS_2_P_STRUC.
// The engine, jackpot, matrix, audio, physics and tasks remain shared.
type PlayerState struct {
	Score, Bonus                                                                    Decimal
	Gear, Speed, Miles, NextJump, NextOffRoad, Position, PositionAvailable, CarPart uint16
	Progress                                                                        [15]bool
}

func (g *Game) SavePlayerState() PlayerState {
	s := PlayerState{Score: g.Score, Bonus: g.Bonus, Gear: g.Gear, Speed: g.Speed, Miles: g.Miles, NextJump: g.NextJump, NextOffRoad: g.NextOffRoad, Position: g.Position, PositionAvailable: g.PositionAvailable, CarPart: g.CarPart}
	for i, n := range []int{9, 10, 11, 12, 13, 14, 22, 23, 24, 25, 50, 51, 52, 53, 54} {
		s.Progress[i] = g.Lights[n]
	}
	return s
}
func (g *Game) LoadPlayerState(s PlayerState) {
	g.Score = s.Score
	g.Bonus = s.Bonus
	g.Gear = s.Gear
	g.Speed = s.Speed
	g.Miles = s.Miles
	g.NextJump = s.NextJump
	g.NextOffRoad = s.NextOffRoad
	g.Position = s.Position
	g.PositionAvailable = s.PositionAvailable
	g.CarPart = s.CarPart
	for i, n := range []int{9, 10, 11, 12, 13, 14, 22, 23, 24, 25, 50, 51, 52, 53, 54} {
		g.Lights[n] = s.Progress[i]
	}
	// DOS loads P_OR_TOTAL/P_TM_TOTAL but never writes them: factory zero.
	g.OffRoadTotal, g.TurboTotal = Decimal{}, Decimal{}
}
func (g *Game) StartPlayers(count int) {
	g.Session.Initialize(count, g.SavePlayerState())
	g.playerText()
	if count > 1 {
		g.beginMatrix("FIRST_NO_OF_PLAYERSTS")
		g.sound("S_ADDPLAYER2")
	}
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
	g.Display.SetPlayers(g.Session.CurrentPlayer, g.Session.PlayerCount, int(g.BallNumber))
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
	g.Display.ShowPlayerBall(g.Score.String())
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
