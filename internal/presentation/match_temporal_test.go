package presentation

import (
	"strings"
	"testing"
)

func TestMatchSourceRetainsMemoryAndSinglePrintSlot(t *testing.T) {
	for table := 1; table <= 4; table++ {
		d := original(t, table)
		for i := range d.Dots {
			d.Dots[i] = true
		}
		want := *d
		want.Text("3"+strings.Repeat(" ", 16), 0, 0, 5)
		d.MatchStart(3)
		if d.Dots != want.Dots {
			t.Fatal("KNACKRUT1 cleared outside LAST_TEXT", table)
		}
		want.Text("*", 7*16, 7, 5)
		d.MatchStep(7, 2)
		if d.Dots != want.Dots {
			t.Fatal("KNACKRUT2 bypassed PRINTTASK", table)
		}
		d.FlushPrint(nil)
		want.Text("2", 2*16, 7, 5)
		if d.Dots != want.Dots {
			t.Fatal("match print coordinates", table)
		}
		d.MatchStep(2, 8)
		// A later handler in this visit replaces match's queued numeric print.
		d.BeginCommand(Command{Op: "_PRINT5", Args: []string{"PLAYERSTEXT", "340"}, Nums: map[int]int{1: 340}})
		expected := *d
		expected.Text(strings.TrimRight(d.SourceText("PLAYERSTEXT"), "\x00"), 8, 1, 5)
		d.FlushPrint(nil)
		if d.Dots != expected.Dots {
			t.Fatal("match did not share PRINTTASK", table)
		}
	}
}
