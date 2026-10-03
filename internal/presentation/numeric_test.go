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
						di := 340
						if centered {
							di -= 2 * (13 - len(tc.visible))
						}
						linkedNumberCommaStores(want, tc.bcd, di, height)
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

// Independent projection of TABLE1.PRG/706a..70e8 (the lost PRINT.ASM).
// LODSW + ADD AL,[SI] scans three BCD bytes per group; CX starts at four.
// DS numeric cells and literal ES VGA stores establish the expectation.
func linkedNumberCommaStores(d *Display, bcd string, di, height int) {
	bcd = strings.Repeat("0", 12-len(bcd)) + bcd
	cx := 4
	for index := 0; index < 9; index += 3 {
		cx--
		if bcd[index] == '0' && bcd[index+1] == '0' && bcd[index+2] == '0' {
			continue
		}
		si := di - (4 - height*168) + int(int16(-2348)) + 2587 - 200
		for ; cx > 0; cx-- {
			for _, store := range [][2]int{{si, 1}, {si + 168, 1}, {si + 1, 0}, {si + 168, 0}} {
				x, y := store[0]%84*2+store[1], store[0]/168-1
				if x >= 0 && x < 160 && y >= 0 && y < 16 {
					d.Dots[y*160+x] = true
				}
			}
			si -= 12
		}
		return
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
			// Independent CODE2 comma stores: restored BX=160 at this
			// SIFFRORRUT site, then SI=BX+0a1bh-200.
			for i := len(s) - 3; i > 0; i -= 3 {
				x := 80 - 8*(len(s)-i) - 1
				for _, p := range [][2]int{{x, 14}, {x, 15}, {x + 1, 14}, {x - 1, 15}} {
					if p[0] >= 0 && p[0] < 160 {
						want.Dots[p[1]*160+p[0]] = true
					}
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
			d.StartCountdown(seconds/10, seconds%10)
			d.StepCountdown("BCD", false, nil)
			d.FlushPrint(func(string) string { return "000000000000" })
			d.StepCountdown("BCD", false, nil)
			d.FlushPrint(func(string) string { return "000000000000" })
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
