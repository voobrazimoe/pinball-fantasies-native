package frontend

import "pinballfantasies/internal/presentation"

func (m *Model) originalCheat(program string) {
	// FANTASIE/TILTRUT, SNAILRUT, BALLSRUT and FAIRPLAYRUT.
	switch program {
	case "QUAKETS":
		m.cheatTiltDisabled = true
	case "SNAILTS":
		m.cheatFastBall = false
	case "BALLSTS":
		m.cheatBalls = 5
	case "FAIRPLAYTS":
		m.cheatTiltDisabled, m.cheatFastBall, m.cheatBalls = false, true, 3
	}
	if session, ok := m.Session.(interface{ MatrixDisplay() *presentation.Display }); ok {
		d := session.MatrixDisplay()
		if m.cheatTimeline == nil {
			current := d.Attract(m.Counter, m.cheatNames(), m.cheatScores())
			if m.gameOverTimeline != nil {
				current = m.gameOverTimeline.Display()
			}
			d.CarryMemory(current)
		}
		m.cheatTimeline = d.StartCheat(program, m.cheatNames(), m.cheatScores())
		m.gameOverTimeline = nil // Accepted replacement owns the sole matrix program.
	}
}

func (m *Model) cheatNames() (names [4]string) {
	for i, score := range m.Scores[m.Selected-1] {
		names[i] = string(score.Name[:])
	}
	return
}
func (m *Model) cheatScores() (scores [4]string) {
	for i, score := range m.Scores[m.Selected-1] {
		scores[i] = score.Digits.String()
	}
	return
}
