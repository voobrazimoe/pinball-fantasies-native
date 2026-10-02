package physics

import (
	"encoding/binary"
	"pinballfantasies/internal/assets"
)

// SHOW configuration; BALLCODE integration/collision algorithms are unchanged.
func DecodeGameshow(data []byte) (*Table, error) {
	p, e := assets.DecodeInitialGameshow(data)
	if e != nil {
		return nil, e
	}
	t := &Table{Initial: p, start: p.State.BallOrigin, flipStride: 55,
		gravity:       [][2]int16{{0, 7}, {4, 9}, {0, 11}, {2, 6}, {6, 10}},
		springInvalid: rectangle{300, 400, 320, 450}, springValid: rectangle{305, 512, 320, 576},
		targets: []rectangle{{159, 122, 180, 131}, {180, 127, 202, 136}, {139, 225, 148, 244}, {135, 245, 144, 264}, {30, 264, 39, 283}, {26, 284, 35, 303}}}
	r := func(o, n int) []byte { return append([]byte(nil), data[o:o+n]...) }
	t.upperForeground = r(0x256f0, 23040)
	t.mask12 = r(0x2b0f0, 23040)
	t.mask11 = r(0x30af0, 23040)
	t.mask22 = r(0x364f0, 23040)
	t.mask13 = r(0x6a3f0, 23040)
	t.mask21 = r(0x6fdf0, 23040)
	t.mask23 = r(0x757f0, 23040)
	t.deltas = r(0xb570, 0x18b60-0xb570)
	t.flipGraphics = r(0x7b1f0, 220)
	for i := range t.Sin {
		t.Sin[i] = int16(binary.BigEndian.Uint16(data[0x1ca40+2*i:]))
	}
	for i := range t.Materials {
		o := 0x1a93d + 16*i
		t.Materials[i] = material{word(data, o), word(data, o+2), word(data, o+4), word(data, o+6), word(data, o+8)}
	}
	t.decodeFlippers(data, 0x1f230, [3]int{0x3bef0, 0x3f4a0, 0x3e510}, r)
	t.bumper = []rectangle{{44, 145, 68, 169}, {74, 201, 98, 226}, {11, 231, 35, 247}}
	t.kicker = []rectangle{{50, 415, 80, 470}, {219, 415, 249, 470}}
	t.levels[1] = []rectangle{{100, 70, 125, 98}, {250, 70, 270, 100}, {255, 85, 277, 120}, {210, 110, 250, 140}, {290, 125, 320, 145}, {140, 125, 170, 170}, {95, 155, 120, 190}, {260, 450, 277, 470}, {20, 450, 50, 470}, {0, 525, 25, 555}}
	t.levels[0] = []rectangle{{125, 70, 150, 100}, {190, 75, 230, 100}, {200, 80, 240, 100}, {290, 105, 320, 125}, {110, 110, 140, 150}}
	return t, nil
}
