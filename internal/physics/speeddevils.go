package physics

import (
	"encoding/binary"
	"pinballfantasies/internal/assets"
)

// DecodeSpeedDevils supplies SDEV's content to the same BALLCODE integrator.
func DecodeSpeedDevils(data []byte) (*Table, error) {
	p, e := assets.DecodeInitialSpeedDevils(data)
	if e != nil {
		return nil, e
	}
	t := &Table{Initial: p, start: p.State.BallOrigin, flipStride: 42,
		// SDEV RAMPTABLE_hi; common TABLE_ANGLE subtracts three from every Y.
		gravity:       [][2]int16{{0, 7}, {0, 12}, {0, 22}, {-1, 7}, {0, 17}, {12, 7}},
		springInvalid: rectangle{300, 480, 320, 500}, springValid: rectangle{305, 512, 320, 576},
		targets: []rectangle{{168, 235, 183, 250}, {183, 246, 199, 258}, {199, 254, 215, 266}, {261, 261, 271, 278}, {271, 278, 278, 294}, {278, 294, 290, 310}}}
	region := func(o, n int) []byte { return append([]byte(nil), data[o:o+n]...) }
	t.upperForeground = region(0x33370, 23040)
	t.mask12 = region(0x38d70, 23040)
	t.mask11 = region(0x3e770, 23040)
	t.mask22 = region(0x44170, 23040)
	t.mask13 = region(0x6d7c0, 23040)
	t.mask21 = region(0x731c0, 20400)
	t.mask23 = region(0x78bc0, 20400)
	t.deltas = region(0xbad0, 54288)
	t.flipGraphics = region(0x7e5c0, 168)
	for i := range t.Sin {
		t.Sin[i] = int16(binary.BigEndian.Uint16(data[0x1d570+2*i:]))
	}
	for i := range t.Materials {
		o := 0x1b00b + 16*i
		t.Materials[i] = material{word(data, o), word(data, o+2), word(data, o+4), word(data, o+6), word(data, o+8)}
	}
	t.decodeFlippers(data, 0x1f820, [3]int{0x49b70, 0x4e110, 0x4c190}, region)
	t.bumper = []rectangle{{52, 193, 76, 217}, {5, 223, 28, 247}, {52, 253, 67, 274}, {5, 283, 28, 307}}
	t.kicker = []rectangle{{50, 415, 80, 470}, {219, 415, 249, 470}}
	t.levels[1] = []rectangle{{160, 40, 195, 70}, {70, 140, 100, 180}, {295, 200, 320, 250}, {230, 210, 280, 250}, {100, 270, 130, 360}, {60, 320, 100, 350}, {260, 450, 277, 470}, {0, 450, 50, 470}}
	t.levels[0] = []rectangle{{70, 115, 100, 140}, {295, 150, 320, 200}, {245, 170, 290, 200}, {60, 280, 80, 320}}
	return t, nil
}
