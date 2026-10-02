package presentation

import (
	"encoding/binary"
	"testing"
)

func TestRuntimeGlyphLiteralAndBoundaryDecoding(t *testing.T) {
	data := make([]byte, 40)
	binary.LittleEndian.PutUint16(data[2:], 8)
	// Synthetic on at x=0, off at x=4, then RET. No original artwork.
	copy(data[8:], []byte{0x88, 0x87, 0, 0, 0x88, 0xa7, 2, 0, 0xc3})
	literal := runtimeGlyphs(data, 0, [][3]int{{0, 0, 0}}, 1, 1, false)
	boundary := runtimeGlyphs(data, 0, [][3]int{{0, 0, 0}}, 1, 1, true)
	if literal["1"][0] != 0x80 || boundary["1"][0] != 0xf0 {
		t.Fatal(literal, boundary)
	}
	data[8] = 0x90
	if runtimeGlyphs(data, 0, [][3]int{{0, 0, 0}}, 1, 1, false) != nil {
		t.Fatal("accepted unsupported drawing store")
	}
	if runtimeGlyphs(data[:2], 0, [][3]int{{0, 0, 0}}, 1, 1, false) != nil {
		t.Fatal("accepted truncated data")
	}
}
