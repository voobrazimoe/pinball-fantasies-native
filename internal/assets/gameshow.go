package assets

import (
	"image"
)

// SHOW STAGE1_1..4, located by original INIT_GFX segment references.
func DecodeGameshow(data []byte) (*Playfield, error) {
	if err := validateLayout("TABLE3.PRG", data); err != nil {
		return nil, err
	}
	p := &Playfield{}
	for _, o := range []int{0x4cb60, 0x52410, 0x5a6b0, 0x634d0} {
		pixels, pal, e := decodeStrip(data, o)
		if e != nil {
			return nil, e
		}
		p.Indices = append(p.Indices, pixels...)
		copy(p.Palette[:], pal)
	}
	return p, nil
}
func DecodeInitialGameshow(data []byte) (*InitialTable, error) {
	p, e := DecodeGameshow(data)
	if e != nil {
		return nil, e
	}
	out := &InitialTable{Playfield: p, State: InitialState{Viewport: image.Rect(0, 259, 320, 576), BallOrigin: image.Pt(299, 530)}, foreground: append([]byte(nil), data[0x1f870:0x1f870+23040]...)}
	// Same linked PUTBALL store layout; Gameshow has its own palette literals.
	for i, o := range ballPaletteOffsets {
		if o != 0 {
			out.Ball[i] = data[o-0xd70]
		}
	}
	return out, nil
}
