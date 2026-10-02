package physics

import "pinballfantasies/internal/settings"

// Configure copies source ramp constants into session ownership. TABLE_ANGLE
// changes only each entry's Y component, never X or another session's ramps.
func (g *Game) Configure(c settings.Config, ramps [][2]int16) {
	g.Settings = c
	g.Configured = true
	g.gravity = append([][2]int16(nil), ramps...)
	if c.Angle == 1 {
		for i := range g.gravity {
			g.gravity[i][1] -= 3
		}
	}
	g.Raster = (g.BottomRaster() + 33) * 16
}
func (g *Game) BottomRaster() int16 { return int16(576 - g.Settings.FieldHeight()) }
func (g *Game) BasePalette(p [768]byte) [768]byte {
	if g.ReferenceMode == 1 {
		return settings.MonoDAC(p)
	}
	return p
}

// LampRGB follows RECALC_LIGHTS -> COLOR_2_BW -> lamp dimming. Average
// after percentage conversion and before halving, exactly as the DOS buffers.
func (g *Game) LampRGB(rgb []byte, on bool) []byte {
	out := make([]byte, len(rgb))
	for i, v := range rgb {
		out[i] = byte(uint16(v) * 162 >> 8)
	}
	if g.ReferenceMode == 1 {
		out = settings.MonoRGB(out)
	}
	for i, v := range out {
		if !on {
			v >>= 1
		}
		out[i] = (v << 2) | (v >> 4)
	}
	return out
}
