package presentation

import (
	"encoding/binary"
	"strconv"
)

// runtimeGlyphs reads literal drawing stores as data, without executing x86.
// The artwork remains in the user's PRG rather than in generated source.
func runtimeGlyphs(data []byte, dsbase int, tables [][3]int, first, last int, transitions bool) map[string][]byte {
	out := make(map[string][]byte)
	for char := first; char <= last; char++ {
		rows := make([]byte, 16)
		edges := make(map[[2]int]bool)
		for _, tab := range tables {
			at := dsbase + tab[0] + 2*char
			if at < 0 || at+2 > len(data) {
				return nil
			}
			q := tab[1] + int(binary.LittleEndian.Uint16(data[at:]))
			for {
				if q < 0 || q >= len(data) {
					return nil
				}
				if data[q] == 0xc3 {
					break
				}
				if q+4 > len(data) || data[q] != 0x88 || (data[q+1] != 0x87 && data[q+1] != 0xa7) {
					return nil
				}
				disp := int(binary.LittleEndian.Uint16(data[q+2:]))
				x, y := (disp%84)*2+tab[2], disp/168
				on := data[q+1] == 0x87
				if y < 16 && x < 8 {
					edges[[2]int{y, x}] = on
					if on {
						rows[y] |= 128 >> x
					}
				}
				q += 4
			}
		}
		if transitions {
			for y := 0; y < 16; y++ {
				rows[y] = 0
				on := false
				for x := 0; x < 8; x++ {
					if edge, ok := edges[[2]int{y, x}]; ok {
						on = edge
					}
					if on {
						rows[y] |= 128 >> x
					}
				}
			}
		}
		out[strconv.Itoa(char)] = rows
	}
	return out
}

func runtimeFonts(table int, data []byte) (map[string][]byte, map[string][]byte) {
	if table < 1 || table > 4 || len(data) == 0 {
		return nil, nil
	}
	i := table - 1
	ds := [4]int{0x19d40, 0x18ee0, 0x18b60, 0x166d0}
	score := [4]int{0xae60, 0xa650, 0xa0f0, 0xb600}
	shift := [4]int{0, 0x90, -0x720, 0x5f0}
	scrollOdd := [4]int{0x7090, 0x6880, 0x6320, 0x7830}
	scrollEven := [4]int{0x7c00, 0x73f0, 0x6e90, 0x83a0}
	return runtimeGlyphs(data, ds[i], [][3]int{{0x5a00 + shift[i], score[i] + 0x1a0, 0}, {0x5c00 + shift[i], score[i] + 0xc40, 1}}, 48, 57, false),
		runtimeGlyphs(data, ds[i], [][3]int{{0x6000 + shift[i], 0x300 + scrollOdd[i], 1}, {0x5e00 + shift[i], 0x300 + scrollEven[i], 0}}, 32, 90, true)
}

// scoreGlyph executes CODE2's linked literal AL/AH stores. Each digit writes
// seven columns by fourteen rows, leaving its eighth column and bottom two
// rows untouched. A decoded bitmap cannot express those retained locations.
func (d *Display) scoreGlyph(char byte, xOffset int) {
	i := d.table - 1
	ds := [4]int{0x19d40, 0x18ee0, 0x18b60, 0x166d0}[i]
	shift := [4]int{0, 0x90, -0x720, 0x5f0}[i]
	code := [4]int{0xae60, 0xa650, 0xa0f0, 0xb600}[i]
	for _, bank := range [][3]int{{0x5a00 + shift, code + 0x1a0, 0}, {0x5c00 + shift, code + 0xc40, 1}} {
		at := ds + bank[0] + 2*int(char)
		q := bank[1] + int(binary.LittleEndian.Uint16(d.data[at:]))
		for d.data[q] != 0xc3 {
			if d.data[q] != 0x88 || d.data[q+1] != 0x87 && d.data[q+1] != 0xa7 {
				panic("unsupported source score glyph store")
			}
			dest := int(binary.LittleEndian.Uint16(d.data[q+2:]))
			x, y := xOffset+dest%84*2+bank[2], dest/168
			if x >= 0 && x < DotWidth && y >= 0 && y < DotHeight {
				d.Dots[y*DotWidth+x] = d.data[q+1] == 0x87
			}
			q += 4
		}
	}
}
