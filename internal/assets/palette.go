package assets

// DACRGB expands the six-bit VGA DAC component, replicating its high bits.
// UNPKLBM shifts the original eight-bit CMAP twice before programming the DAC.
func DACRGB(v byte) byte { v &= 63; return v<<2 | v>>4 }
func VGAPalette(raw [768]byte) [768]byte {
	for i, v := range raw {
		raw[i] = DACRGB(v >> 2)
	}
	return raw
}
