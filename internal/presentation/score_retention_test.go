package presentation

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

// Independent byte-address VGA projection of the linked CODE2 font stores.
func oracleScoreGlyph(d *Display, digit byte, di int) {
	i := d.table - 1
	ds := [4]int{0x19d40, 0x18ee0, 0x18b60, 0x166d0}[i]
	delta := [4]int{0, 0x90, -0x720, 0x5f0}[i]
	code := [4]int{0xae60, 0xa650, 0xa0f0, 0xb600}[i]
	for _, bank := range [][3]int{{0x5a00 + delta, code + 0x1a0, 0}, {0x5c00 + delta, code + 0xc40, 1}} {
		pc := bank[1] + int(binary.LittleEndian.Uint16(d.data[ds+bank[0]+int(digit)*2:]))
		for d.data[pc] != 0xc3 {
			addr := di + int(binary.LittleEndian.Uint16(d.data[pc+2:]))
			x, y := addr%84*2+bank[2], addr/168-1
			if x >= 0 && x < 160 && y >= 0 && y < 16 {
				d.Dots[y*160+x] = d.data[pc+1] == 0x87
			}
			pc += 4
		}
	}
}

func TestCODE2RetainedCellsCacheAndWidthTransitions(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			d := original(t, table)
			want := *d
			for i := range d.Dots {
				d.Dots[i] = i%3 == 0
				want.Dots[i] = i%3 == 0
			}
			cache := [12]byte{}
			for i := range cache {
				cache[i] = 0x12
			}
			commas := 0
			for _, score := range []string{"000000000000", "1000", "1000", "10000", "1234567", "123", "0", "1234567"} {
				visible := strings.TrimLeft(score, "0")
				if visible == "" {
					visible = "0"
				}
				for index, digit := range []byte(visible) {
					if cache[index] != digit-'0' {
						oracleScoreGlyph(&want, digit, 248-4*len(visible)+4*index)
						cache[index] = digit - '0'
					}
				}
				groups := (len(visible) - 1) / 3
				if groups != 0 && groups != commas {
					commas = groups
					for n := 1; n <= groups; n++ {
						// CODE2: plane4 [SI] and [SI+168], plane1 [SI+1]
						// and [SI+168], where SI=BX+0a1bh-200-12*(n-1).
						addr := 200 + 0x0a1b - 200 - 12*(n-1)
						for _, store := range [][2]int{{addr, 1}, {addr + 168, 1}, {addr + 1, 0}, {addr + 168, 0}} {
							x, y := store[0]%84*2+store[1], store[0]/168-1
							if x >= 0 && x < 160 && y >= 0 && y < 16 {
								want.Dots[y*160+x] = true
							}
						}
					}
				}
				d.Score(score)
				if d.Dots != want.Dots {
					t.Fatal("CODE2 literal retained pixels", score)
				}
			}
			// Interruptions erase pixels without invalidating oldbuf. DO_SPEC_MATRIX
			// invokes UPDAT_SCORE, so ordinary score restoration redraws changed cells.
			d.Clear()
			before := d.Dots
			d.Score("1234567")
			if d.Dots != before {
				t.Fatal("unchanged digits redrawn after direct VGA clear")
			}
			d.BeginCommand(Command{Op: "_WAIT", Nums: map[int]int{0: 1}})
			d.Score("1234567")
			if d.Dots == before {
				t.Fatal("source command failed to invalidate cached score")
			}
		})
	}
}
