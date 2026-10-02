package assets

import (
	"crypto/sha256"
	"fmt"
	"image"
)

const Table3SHA256 = "da83ef5a7a471e6a6ad759126907076c81e92ffde6dec8e3de8e6052c6a98858"

// SHOW STAGE1_1..4, located by original INIT_GFX segment references.
func DecodeGameshow(data []byte) (*Playfield, error) {
	if fmt.Sprintf("%x", sha256.Sum256(data)) != Table3SHA256 {
		return nil, fmt.Errorf("TABLE3.PRG differs from inventoried Gameshow build")
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
