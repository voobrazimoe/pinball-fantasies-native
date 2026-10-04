package assets

import (
	"image"
)

// STONES BANH=576. STAGE4 has additional rows in its linked PBM:
// 320x1219, independently decoded through the complete BODY extent.
func DecodeStones(data []byte) (*Playfield, error) {
	if err := validateLayout("TABLE4.PRG", data); err != nil {
		return nil, err
	}
	p := &Playfield{}
	for i, o := range []int{0x4bc10, 0x54a00, 0x5da70, 0x66e20} {
		h := 144
		if i == 3 {
			h = 1219
		}
		pixels, pal, e := decodeStripHeight(data, o, h)
		if e != nil {
			return nil, e
		}
		p.Indices = append(p.Indices, pixels[:320*144]...)
		copy(p.Palette[:], pal)
	}
	return p, nil
}
func DecodeInitialStones(data []byte) (*InitialTable, error) {
	p, e := DecodeStones(data)
	if e != nil {
		return nil, e
	}
	out := &InitialTable{Playfield: p, State: InitialState{Viewport: image.Rect(0, 259, 320, 576), BallOrigin: image.Pt(297, 530)}, foreground: append([]byte(nil), data[0x28f30:0x28f30+23040]...)}
	for i, o := range ballPaletteOffsets {
		if o != 0 {
			out.Ball[i] = data[o+0x7a0]
		}
	}
	return out, nil
}
