package frontend

import (
	"image"
	"image/color"
	"pinballfantasies/internal/assets"
)

// PFTASK's counter runs only in the frontend. Its INFO branch clears three
// scanlines per callback, writes one BIOS character per callback, then holds;
// LOGGAN restores the saved original 112x90 sidebar one line per callback.
func (v *View) sidebar(out *image.RGBA, tick int) {
	v.sidebarInfo(out, tick, v.Art.SidebarInfo)
}
func (v *View) sidebarInfo(out *image.RGBA, tick int, info string) {
	tick %= 1200
	if tick < 480 {
		return
	}
	p := v.Art.Logo.Palette
	background := color.RGBA{p[6], p[7], p[8], 255}
	ink := color.RGBA{p[0], p[1], p[2], 255}
	clear := func(first, count int) {
		for y := first; y < first+count; y++ {
			for x := 8; x < 120; x++ {
				out.SetRGBA(x, y, background)
			}
		}
	}
	if tick < 510 {
		clear(95, (tick-480+1)*3)
		return
	}
	clear(95, 90)
	chars := 120
	if tick < 630 {
		chars = tick - 509
	}
	for i := 0; i < chars; i++ {
		ch := info[i]
		for row := 0; row < 8; row++ {
			bits := assets.SidebarFont[int(ch)*8+row]
			for col := 0; col < 8; col++ {
				c := background
				if bits&(128>>uint(col)) != 0 {
					c = ink
				}
				out.SetRGBA(16+i%12*8+col, 97+i/12*9+row, c)
			}
		}
	}
	if tick >= 1110 {
		first := 184 - (tick - 1110)
		for y := first; y <= 184; y++ {
			for x := 16; x < 128; x++ {
				idx := int(v.Art.Logo.Indices[y*v.Art.Logo.Width+x]) * 3
				out.SetRGBA(x, y, color.RGBA{p[idx], p[idx+1], p[idx+2], 255})
			}
		}
	}
}
