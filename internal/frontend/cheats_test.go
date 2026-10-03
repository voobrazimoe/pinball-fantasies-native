package frontend

import (
	"pinballfantasies/internal/presentation"
	"testing"
)

func typeCheat(t *testing.T, m *Model, word string) {
	t.Helper()
	letters := "QWERTYUIOPASDFGHJKLZXCVBNM"
	scans := []Key{16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 30, 31, 32, 33, 34, 35, 36, 37, 38, 44, 45, 46, 47, 48, 49, 50}
	for _, char := range word {
		if char == ' ' {
			key(t, m, Space)
			continue
		}
		for i, letter := range letters {
			if letter == char {
				key(t, m, scans[i])
				break
			}
		}
	}
}

func TestOriginalCheatsOnlyOnLoadedTableAttract(t *testing.T) {
	// MAIN calls CHECKCHEAT only in DEMOMODE; WHEN_NEW_GAME_RESET
	// resets AFTER_CHEAT but does not reset SHIFTKEYS/TILTDISABLED/NO_OF_BALLS.
	for table := 1; table <= 4; table++ {
		r := hotseatRuntime(t)
		m := r.Model
		key(t, m, Key(int(F1)+table-1))
		typeCheat(t, m, "EARTHQUAKE")
		if !m.cheatTiltDisabled || m.cheatTimeline == nil {
			t.Fatal("TILTRUT", table)
		}
		typeCheat(t, m, "EXTRA BALLS")
		if m.cheatBalls != 5 {
			t.Fatal("BALLSRUT writes five, despite message", table)
		}
		typeCheat(t, m, "FAIR PLAY")
		if m.cheatTiltDisabled || !m.cheatFastBall || m.cheatBalls != 3 {
			t.Fatal("FAIRPLAYRUT", table)
		}
		typeCheat(t, m, "SNAIL")
		if m.cheatFastBall {
			t.Fatal("SNAILRUT sets SHIFTKEYS bit2", table)
		}
		typeCheat(t, m, "EXTRA BALLS")
		typeCheat(t, m, "EARTHQUAKE")
		previous := m.cheatTimeline.Display().Dots
		key(t, m, F1)
		if m.Mode != Playing || m.cheatTimeline != nil || m.Session.(interface{ BaseBalls() byte }).BaseBalls() != 5 {
			t.Fatal("NEW_GAME cheated globals", table)
		}
		if m.Session.(interface{ MatrixDisplay() *presentation.Display }).MatrixDisplay().Dots != previous {
			t.Fatal("start discarded cheat VGA memory", table)
		}
		// CHEAT's letters don't pause, tilt or select players. During play,
		// no CHECKCHEAT call is permitted by the original MAIN branch.
		typeCheat(t, m, "CHEAT")
		if m.cheatTimeline != nil {
			t.Fatal("in-game cheat accepted", table)
		}
	}
}
