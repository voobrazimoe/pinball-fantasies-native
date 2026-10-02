package presentation

// HighCRTC is SET_360X350's final register set, after the horizontal override
// and SETSPLIT. The 336-pixel pitch is register 13h*8, not the display width.
var HighCRTC = map[byte]byte{
	0x00: 0x6b, 0x01: 79, 0x02: 0x5a, 0x03: 0x8e, 0x04: 0x5b, 0x05: 0x87,
	0x06: 0xbf, 0x07: 0x1f, 0x09: 0x00, 0x10: 0x83, 0x11: 0xa5, 0x12: 0x5d,
	0x13: 42, 0x14: 0x0f, 0x15: 0x63, 0x16: 0xba, 0x17: 0xe3, 0x18: 0x3c,
}

// SETSPLIT replaces bit 4 of overflow for line-compare bit 8 (316).
func CRTCHeight(r map[byte]byte) int {
	return 1 + int(r[0x12]) + int(r[0x07]&2)*128 + int(r[0x07]&64)*8
}

// GeometricPixels retains the source ball's equal 15x15 occupied diameters.
// It is a documented artwork-geometry policy, not a claim about CRT calibration.
// DOSBox's raw 640x350 surface uses two clock pixels per indexed source pixel.
const GeometricPixelAspectNumerator = 1
const GeometricPixelAspectDenominator = 1

// NormalCRTC records SET240's ten writes and SETSPLIT's line compare.
// Unlike runtime HI_RES, build-time HIRES is false: bit7 of09h stays set.
var NormalCRTC = map[byte]byte{0x06: 0x0d, 0x07: 0x3e, 0x09: 0x80, 0x10: 0xea, 0x11: 0x0c, 0x12: 0xdf, 0x14: 0, 0x15: 0xea, 0x16: 6, 0x17: 0xe3, 0x18: 0x9d}
