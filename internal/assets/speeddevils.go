package assets

import (
	"crypto/sha256"
	"fmt"
	"image"
)

const Table2SHA256 = "6689dcef5fd051998bab990b5d243614c7dae2dcdcab9bffbe3c1936a76504b5"

// SDEV BANH=576 and shared INIT_GFX's four 320x144 PBM strips. Static linked
// INIT_GFX references locate these original assets; no executable code runs.
func DecodeSpeedDevils(data []byte) (*Playfield, error) {
	if fmt.Sprintf("%x", sha256.Sum256(data)) != Table2SHA256 {
		return nil, fmt.Errorf("TABLE2.PRG differs from inventoried Speed Devils build")
	}
	p := &Playfield{}
	for _, off := range []int{0x50730, 0x583f0, 0x60030, 0x67b00} {
		pixels, palette, e := decodeStrip(data, off)
		if e != nil {
			return nil, e
		}
		p.Indices = append(p.Indices, pixels...)
		copy(p.Palette[:], palette)
	}
	return p, nil
}

func DecodeInitialSpeedDevils(data []byte) (*InitialTable, error) {
	p, e := DecodeSpeedDevils(data)
	if e != nil {
		return nil, e
	}
	// SDEV STARTX=310-8+3; STARTY=543-8; SETBALL subtracts five.
	out := &InitialTable{Playfield: p, State: InitialState{Viewport: image.Rect(0, 259, 320, 576), BallOrigin: image.Pt(300, 530)}, foreground: append([]byte(nil), data[0x2d4f0:0x2d4f0+23040]...)}
	// Same external unrolled PUTBALL sprite, linked 0x810 bytes earlier in TABLE2.
	// The location map reads only its literal palette bytes.
	for i, off := range ballPaletteOffsets {
		if off != 0 {
			out.Ball[i] = data[off-0x810]
		}
	}
	return out, nil
}
