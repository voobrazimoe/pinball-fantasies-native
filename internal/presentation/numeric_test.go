package presentation

import (
	"fmt"
	"strings"
	"testing"
)

func TestAllTablesSourceNumericFormats(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for _, height := range []int{5, 8, 11, 13} {
			for _, centered := range []bool{false, true} {
				for _, tc := range []struct{ bcd, visible string }{
					{"000000000002", "2"}, {"000000000010", "10"}, {"000000000100", "100"},
					{"123456789012", "123456789012"}, {"000000000000", ""},
				} {
					t.Run(fmt.Sprintf("table%d-font%d-center%t-%s", table, height, centered, tc.bcd), func(t *testing.T) {
						d := original(t, table)
						op := fmt.Sprintf("_PRINT%d_NUMBER", height)
						if centered {
							op += "_CENT"
						}
						d.BeginCommand(Command{Op: op, Args: []string{"TESTBCD", "340"}, Nums: map[int]int{1: 340}})
						d.Visit(0, 0, func(string) string { return tc.bcd })
						want := original(t, table)
						// Source DI=340 means x=8,y=1. Each skipped digit advances
						// four VGA bytes (eight dots). CENT subtracts 2 bytes per
						// SCASB visit, including the first nonzero BCD byte.
						x := 8 + 8*(12-len(tc.visible))
						if centered {
							x -= 4 * (13 - len(tc.visible))
						}
						want.Text(tc.visible, x, 1, height)
						if d.Dots != want.Dots {
							t.Fatal("source BCD suppression/alignment differs")
						}
						if !centered {
							d.Clear()
							d.Number(tc.bcd, 8, 1, height)
							if d.Dots != want.Dots {
								t.Fatal("table-local/countdown numeric path bypasses PRINT_NUMBER")
							}
						}
					})
				}
			}
		}
	}
}

func TestNumberOpcodeUsesGeneratedScoreRules(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for _, value := range []string{"000000000002", "000000000010", "000000000100", "123456789012", "000000000000"} {
			d := original(t, table)
			d.Begin("_NUMBER", []string{"BCD"})
			d.Visit(0, 0, func(string) string { return value })
			want := original(t, table)
			s := strings.TrimLeft(value, "0")
			if s == "" {
				s = "0"
			}
			want.GlyphText(s, 80-8*len(s), 0, want.Content.ScoreFont)
			// Independent CODE2 comma positions (including large BCD scores).
			for i := len(s) - 3; i > 0; i -= 3 {
				x := 80 - 8*(len(s)-i) - 1
				for y := 13; y < 15; y++ {
					if x >= 0 && x < 160 {
						want.Dots[y*160+x] = true
					}
				}
				if x > 0 && x < 160 {
					want.Dots[15*160+x-1] = true
				}
			}
			if d.Dots != want.Dots {
				t.Fatal("_NUMBER CODE2 rules changed", table, value)
			}
		}
	}
}

func TestOrdinarySourceTextsKeepEncodedZeroes(t *testing.T) {
	for table := 1; table <= 4; table++ {
		d := original(t, table)
		d.Content.Texts["TESTTEXT"] = []byte{'7', '7', '9', 0} // intentional source "002"
		d.BeginCommand(Command{Op: "_PRINT5", Args: []string{"TESTTEXT", "340"}, Nums: map[int]int{1: 340}})
		d.Visit(0, 0, nil)
		want := original(t, table)
		want.Text("002", 8, 1, 5)
		if d.Dots != want.Dots {
			t.Fatal("ordinary source text was globally trimmed")
		}
	}
}

func TestCountdownSourceZeroAndSeconds(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for _, seconds := range []int{0, 2, 10, 25, 50} {
			d := original(t, table)
			d.Countdown("000000000000", seconds)
			want := original(t, table)
			// FANTASIE countdown blanks SEC_ASC's zero tens; zero numeric value
			// follows PRINT_NUMBER and contributes no large-font zero glyph.
			want.Text(fmt.Sprintf("%2d", seconds), 144, 2, 11)
			if d.Dots != want.Dots {
				t.Fatal("countdown bypassed source formatting", table, seconds)
			}
		}
	}
}
