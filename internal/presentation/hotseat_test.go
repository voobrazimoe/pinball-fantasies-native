package presentation

import (
	"strings"
	"testing"
)

func TestHotseatSourcePlayerDigits(t *testing.T) {
	for table := 1; table <= 4; table++ {
		d := original(t, table)
		d.SetPlayers(8, 8, 3)
		if !strings.Contains(d.SourceText("PLAYERSTEXT"), "8") || !strings.Contains(d.SourceText("BALLSTEXT"), "3") || !strings.Contains(d.SourceText("NO_OF_PLAYERS_TEXT"), "8") {
			t.Fatalf("table %d player/ball/count text", table)
		}
		d.Begin("_SHOW_SCORE", nil)
		d.Visit(0, 0, func(string) string { return "000000012340" })
		want := original(t, table)
		want.Text("PLAYER 8", 0, 1, 5)
		want.Score("000000012340")
		if d.Dots != want.Dots {
			t.Fatalf("table %d current-player score panel", table)
		}
	}
}
