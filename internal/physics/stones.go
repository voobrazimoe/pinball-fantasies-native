package physics

import (
	"encoding/binary"
	"pinballfantasies/internal/assets"
)

func DecodeStones(data []byte) (*Table, error) {
	p, e := assets.DecodeInitialStones(data)
	if e != nil {
		return nil, e
	}
	t := &Table{Initial: p, start: p.State.BallOrigin, flipStride: 30,
		gravity:     [][2]int16{{0, 10}, {-10, 5}, {0, -10}, {5, 0}, {5, 15}, {-10, 12}, {2, 15}, {-8, 12}, {3, 10}, {4, 13}, {7, 10}},
		springValid: rectangle{305, 544, 320, 576}, springInvalid: rectangle{300, 520, 315, 540},
		targets: []rectangle{{123, 226, 143, 243}, {143, 230, 164, 243}, {70, 234, 81, 241}, {217, 234, 229, 241}, {11, 277, 38, 294}, {11, 294, 38, 310}, {11, 310, 38, 325}, {272, 303, 286, 323}, {272, 323, 286, 343}}}
	r := func(o, n int) []byte { return append([]byte(nil), data[o:o+n]...) }
	t.upperForeground = r(0x2edb0, 23040)
	t.mask12 = r(0x347b0, 23040)
	t.mask11 = r(0x3a1b0, 23040)
	t.mask22 = r(0x3fbb0, 23040)
	t.mask13 = r(0x6e870, 23040)
	t.mask21 = r(0x74270, 23040)
	t.mask23 = r(0x79c70, 23040)
	t.deltas = r(0xca80, 0x166d0-0xca80)
	t.flipGraphics = r(0x7f670, 120)
	for i := range t.Sin {
		t.Sin[i] = int16(binary.BigEndian.Uint16(data[0x1b2c0+2*i:]))
	}
	for i := range t.Materials {
		o := 0x18e77 + 16*i
		t.Materials[i] = material{word(data, o), word(data, o+2), word(data, o+4), word(data, o+6), word(data, o+8)}
	}
	t.decodeFlippers(data, 0x1da30, [3]int{0x455b0, 0x47bd0, 0}, r)
	t.bumper = []rectangle{{172, 88, 185, 112}, {231, 94, 255, 118}, {190, 134, 214, 158}}
	t.kicker = []rectangle{{50, 415, 80, 470}, {219, 415, 249, 470}}
	t.levels[1] = []rectangle{{16, 8, 42, 45}, {90, 80, 115, 120}, {265, 155, 295, 185}, {74, 169, 88, 181}, {74, 181, 104, 193}, {185, 200, 215, 240}, {300, 200, 320, 250}, {260, 450, 277, 470}, {20, 450, 50, 470}, {1, 526, 20, 576}}
	t.levels[0] = []rectangle{{42, 8, 66, 38}, {136, 36, 156, 60}, {91, 38, 133, 55}, {138, 140, 157, 160}, {190, 165, 225, 200}, {300, 170, 320, 200}}
	return t, nil
}
