package gameshow

import (
	"pinballfantasies/internal/presentation"
	"testing"
)

// Exercise the actual source bonus opcode and the table's live BCD reader.
func TestBonusCounterUsesSourcePrintNumber(t *testing.T) {
	for _, value := range []uint16{0, 2, 10, 100} {
		g := game(t)
		g.Skills = value
		visible := map[uint16]string{0: "", 2: "2", 10: "10", 100: "100"}[value]
		found := false
		for _, c := range g.Display.Content.Commands {
			if c.Op != "_PRINT5_NUMBER" || c.Args[0] != "CYCLONECOUNTERBCD" {
				continue
			}
			found = true
			g.Display.Clear()
			g.Display.BeginCommand(c)
			g.Display.Visit(0, 0, g.matrixNumber)
			want := game(t).Display
			x, y := presentation.PositionValue(c.Num(1))
			want.Text(visible, x+8*(12-len(visible)), y, 5)
			if g.Display.Dots != want.Dots {
				t.Fatal("source bonus counter has extra leading digits", value)
			}
		}
		if !found {
			t.Fatal("source bonus counter opcode absent")
		}
	}
}

func TestJackValueNumericReader(t *testing.T) {
	g := game(t)
	g.Jackpot = number(123456789012)
	if got := g.matrixNumber("JACKVALUE"); got != "123456789012" {
		t.Fatal("source JACKVALUE bypassed live BCD reader", got)
	}
}
