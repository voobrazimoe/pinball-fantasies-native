package presentation

import "encoding/binary"

// scrollStores decodes the literal AL/AH drawing stores used by linked
// SCROLLE. These are moving run boundaries, not a freshly cleared bitmap.
// A replacement resets the base addresses but inherits plane selectors and
// phase, so integrating the glyph once and shifting it is insufficient.
func (d *Display) scrollStores() {
	i := d.table - 1
	ds := [4]int{0x19d40, 0x18ee0, 0x18b60, 0x166d0}[i]
	shift := [4]int{0, 0x90, -0x720, 0x5f0}[i]
	odd := [4]int{0x7090, 0x6880, 0x6320, 0x7830}[i] + 0x300
	even := [4]int{0x7c00, 0x73f0, 0x6e90, 0x83a0}[i] + 0x300
	text := d.Content.Texts[d.args[0]]
	// SCROLLE visits the odd-glyph code first, then the even-glyph code.
	for _, plane := range []int{1, 0} {
		tab, code := 0x5e00+shift, even
		if plane == 1 {
			tab, code = 0x6000+shift, odd
		}
		for char := 0; char < 21; char++ {
			at := ds + tab + 2*int(text[d.scrollChar+char])
			q := code + int(binary.LittleEndian.Uint16(d.data[at:]))
			for d.data[q] != 0xc3 {
				if d.data[q] != 0x88 || d.data[q+1] != 0x87 && d.data[q+1] != 0xa7 {
					panic("unsupported source scroll glyph store")
				}
				dest := d.scrollBase[plane] + char*4 + int(binary.LittleEndian.Uint16(d.data[q+2:]))
				x, y := dest%84*2+d.scrollPlane[plane], dest/168-1
				if x >= 0 && x < DotWidth && y >= 0 && y < DotHeight {
					d.Dots[y*DotWidth+x] = d.data[q+1] == 0x87
				}
				q += 4
			}
		}
	}
}
