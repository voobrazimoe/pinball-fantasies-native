// Package physics ports the original Party Land ball subsystem as integer Go.
package physics

import (
	"encoding/binary"
	"image"
	"pinballfantasies/internal/assets"
)

type rectangle struct{ X1, Y1, X2, Y2 int16 }

func (r rectangle) contains(x, y int16) bool {
	return uint16(x) >= uint16(r.X1) && uint16(x) <= uint16(r.X2) && uint16(y) >= uint16(r.Y1) && uint16(y) <= uint16(r.Y2)
}

type material struct{ WallFriction, BallFriction, Bounce, MinSpeed, MaxAngle int16 }
type Flipper struct {
	Kind                                                                                uint8
	Left, Top, Words, Height                                                            int16
	Bounds                                                                              rectangle
	CenterX, CenterY, PowerZone                                                         int16
	Speed, Angle, Frame, Frames, MaxAngle, AccelerationUp, AccelerationDown, MaxSpeedUp int16
	masks                                                                               []byte
	stride                                                                              int
}
type Table struct {
	start                                          image.Point
	gravity                                        [][2]int16
	targets                                        []rectangle
	springInvalid, springValid                     rectangle
	flipStride                                     int
	Initial                                        *assets.InitialTable
	Sin                                            [2560]int16
	Materials                                      [8]material
	Flippers                                       [3]Flipper
	mask11, mask12, mask22, mask13, mask21, mask23 []byte
	levels                                         [2][]rectangle
	bumper, kicker                                 []rectangle
	upperForeground, deltas, flipGraphics          []byte
	deltaBase, deltaSize, deltaMax                 [3]int
}

func DecodePartyLand(data []byte) (*Table, error) {
	p, err := assets.DecodeInitialPartyLand(data)
	if err != nil {
		return nil, err
	}
	t := &Table{Initial: p, start: p.State.BallOrigin, flipStride: 437,
		gravity:       [][2]int16{{0, 7}, {2, 11}, {-2, 11}, {-4, 13}},
		targets:       []rectangle{{130, 196, 146, 204}, {147, 277, 156, 293}, {152, 293, 163, 311}, {158, 311, 168, 329}},
		springInvalid: rectangle{305, 455, 320, 540}, springValid: rectangle{308, 540, 320, 576}}
	copyRegion := func(o, n int) []byte { return append([]byte(nil), data[o:o+n]...) }
	t.upperForeground = copyRegion(0x35f30, 23040)
	t.deltas = copyRegion(0xc2e0, 55888)
	t.flipGraphics = copyRegion(0x82930, 1748)
	t.mask11 = copyRegion(0x41330, 23040)
	t.mask12 = copyRegion(0x3b930, 23040)
	t.mask22 = copyRegion(0x46d30, 23040)
	t.mask13 = copyRegion(0x71b30, 23040)
	t.mask21 = copyRegion(0x77530, 20400)
	t.mask23 = copyRegion(0x7cf30, 20400)
	for i := range t.Sin {
		t.Sin[i] = int16(binary.BigEndian.Uint16(data[0x1e340+2*i:]))
	}
	for i := range t.Materials {
		o := 0x1c05d + i*16
		t.Materials[i] = material{word(data, o), word(data, o+2), word(data, o+4), word(data, o+6), word(data, o+8)}
	}
	t.decodeFlippers(data, 0x20690, [3]int{0x4c730, 0x4fe10, 0x4ed50}, copyRegion)
	t.bumper = []rectangle{{211, 252, 235, 276}, {268, 263, 292, 287}, {185, 283, 209, 307}}
	t.kicker = []rectangle{{50, 415, 80, 470}, {219, 415, 249, 470}}
	t.levels[1] = []rectangle{{10, 100, 40, 200}, {300, 140, 320, 160}, {275, 165, 300, 220}, {60, 185, 100, 220}, {195, 175, 215, 200}, {300, 240, 320, 340}, {20, 240, 50, 400}, {300, 400, 320, 500}, {260, 450, 277, 470}, {0, 450, 50, 470}}
	t.levels[0] = []rectangle{{215, 160, 235, 190}, {50, 165, 85, 185}, {300, 210, 320, 240}, {3, 245, 22, 270}, {300, 360, 320, 400}}
	return t, nil
}

func (t *Table) decodeFlippers(data []byte, descriptor int, frameOffsets [3]int, copyRegion func(int, int) []byte) {
	// INIT_FLIPPERS uses the generated collision-only frames, then ORMASKFLIP
	// adds the table's original collision bytes into every frame.
	for i := range t.Flippers {
		o := descriptor + i*60
		f := &t.Flippers[i]
		f.Kind = data[o]
		if f.Kind == 0 {
			break
		}
		f.Left = word(data, o+2)
		f.Top = word(data, o+4)
		f.Words = word(data, o+6)
		f.Height = word(data, o+8)
		f.Bounds = rectangle{word(data, o+10), word(data, o+14), word(data, o+12), word(data, o+16)}
		f.CenterX = word(data, o+18)
		f.CenterY = word(data, o+20)
		f.PowerZone = word(data, o+22)
		f.Speed = word(data, o+26)
		f.Angle = word(data, o+28)
		f.Frame = word(data, o+30)
		f.Frames = word(data, o+32)
		f.MaxAngle = word(data, o+34)
		f.AccelerationUp = word(data, o+36)
		f.AccelerationDown = word(data, o+38)
		f.MaxSpeedUp = word(data, o+40)
		f.stride = int(f.Height * f.Words * 2)
		t.deltaBase[i] = int(uint16(word(data, o+54))) + 8*i
		t.deltaSize[i] = int(uint16(word(data, o+56)))
		t.deltaMax[i] = int(uint16(word(data, o+48))) + 8*i
		f.masks = copyRegion(frameOffsets[i], f.stride*(int(f.Frames)+1))
		for frame := 0; frame <= int(f.Frames); frame++ {
			for y := 0; y < int(f.Height); y++ {
				for x := 0; x < int(f.Words)*2; x++ {
					f.masks[frame*f.stride+y*int(f.Words)*2+x] |= t.mask12[(int(f.Top)+y)*40+int(f.Left)/8+x]
				}
			}
		}
	}
}

func word(data []byte, o int) int16 { return int16(binary.LittleEndian.Uint16(data[o:])) }

// The collision/material bitmaps are MSB first, 40 bytes per original row.
// Preserve the original linear byte addressing, including samples past x=319
// spilling into the following row. Reads outside the supplied map are empty;
// native code never reads unrelated DOS segment memory.
func bit(m []byte, x, y int16) bool {
	i := int(y)*40 + (int(x) >> 3)
	if i < 0 || i >= len(m) {
		return false
	}
	return m[i]&(128>>uint(x&7)) != 0
}
