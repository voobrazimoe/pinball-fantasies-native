package gamepad

import (
	"image"
	"image/color"
	"image/draw"
)

// Authored 5x7 UI lettering. No DOS/game fonts or external icon assets.
var letters = map[byte]string{
	'A': "01110/10001/10001/11111/10001/10001/10001", 'B': "11110/10001/10001/11110/10001/10001/11110",
	'C': "01111/10000/10000/10000/10000/10000/01111", 'D': "11110/10001/10001/10001/10001/10001/11110",
	'E': "11111/10000/10000/11110/10000/10000/11111", 'F': "11111/10000/10000/11110/10000/10000/10000",
	'G': "01111/10000/10000/10111/10001/10001/01111", 'H': "10001/10001/10001/11111/10001/10001/10001",
	'I': "11111/00100/00100/00100/00100/00100/11111", 'J': "00111/00010/00010/00010/10010/10010/01100",
	'K': "10001/10010/10100/11000/10100/10010/10001", 'L': "10000/10000/10000/10000/10000/10000/11111",
	'M': "10001/11011/10101/10101/10001/10001/10001", 'N': "10001/11001/10101/10011/10001/10001/10001",
	'O': "01110/10001/10001/10001/10001/10001/01110", 'P': "11110/10001/10001/11110/10000/10000/10000",
	'Q': "01110/10001/10001/10001/10101/10010/01101", 'R': "11110/10001/10001/11110/10100/10010/10001",
	'S': "01111/10000/10000/01110/00001/00001/11110", 'T': "11111/00100/00100/00100/00100/00100/00100",
	'U': "10001/10001/10001/10001/10001/10001/01110", 'V': "10001/10001/10001/10001/10001/01010/00100",
	'W': "10001/10001/10001/10101/10101/10101/01010", 'X': "10001/10001/01010/00100/01010/10001/10001",
	'Y': "10001/10001/01010/00100/00100/00100/00100", 'Z': "11111/00001/00010/00100/01000/10000/11111",
	'-': "00000/00000/00000/11111/00000/00000/00000", '/': "00001/00001/00010/00100/01000/10000/10000",
	'\'': "00100/00100/00000/00000/00000/00000/00000", '1': "00100/01100/00100/00100/00100/00100/01110",
	'2': "01110/10001/00001/00010/00100/01000/11111",
}
var ink = color.RGBA{240, 245, 250, 255}

func text(dst *image.RGBA, s string, x, y int, c color.RGBA) {
	for i := 0; i < len(s); i++ {
		row, col := 0, 0
		for _, bit := range letters[s[i]] {
			if bit == '/' {
				row++
				col = 0
				continue
			}
			if bit == '1' {
				dst.SetRGBA(x+col, y+row, c)
			}
			col++
		}
		x += 6
	}
}
func line(dst *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		dst.SetRGBA(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e := 2 * err
		if e >= dy {
			err += dy
			x0 += sx
		}
		if e <= dx {
			err += dx
			y0 += sy
		}
	}
}
func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
func glyph(dst *image.RGBA, p *Input, button, x, y int) {
	c := ink
	if button >= 0 && button < 4 {
		g := p.Glyph(button)
		switch g {
		case Cross:
			c = color.RGBA{130, 190, 255, 255}
			line(dst, x+2, y+2, x+10, y+10, c)
			line(dst, x+10, y+2, x+2, y+10, c)
		case Circle:
			c = color.RGBA{255, 150, 160, 255}
			for dx := -6; dx <= 6; dx++ {
				for dy := -6; dy <= 6; dy++ {
					r := dx*dx + dy*dy
					if r >= 25 && r <= 40 {
						dst.SetRGBA(x+6+dx, y+6+dy, c)
					}
				}
			}
		case Square:
			c = color.RGBA{230, 155, 230, 255}
			line(dst, x+1, y+1, x+11, y+1, c)
			line(dst, x+11, y+1, x+11, y+11, c)
			line(dst, x+11, y+11, x+1, y+11, c)
			line(dst, x+1, y+11, x+1, y+1, c)
		case Triangle:
			c = color.RGBA{125, 230, 200, 255}
			line(dst, x+6, y, x, y+11, c)
			line(dst, x, y+11, x+12, y+11, c)
			line(dst, x+12, y+11, x+6, y, c)
		default:
			label := map[Glyph]string{LetterA: "A", LetterB: "B", LetterX: "X", LetterY: "Y"}[g]
			colors := map[Glyph]color.RGBA{LetterA: {140, 225, 150, 255}, LetterB: {255, 150, 150, 255}, LetterX: {145, 190, 255, 255}, LetterY: {245, 220, 140, 255}}
			c = colors[g]
			line(dst, x, y, x+12, y, c)
			line(dst, x, y+12, x+12, y+12, c)
			line(dst, x, y, x, y+12, c)
			line(dst, x+12, y, x+12, y+12, c)
			text(dst, label, x+4, y+3, c)
		}
		return
	}
	switch button {
	case LT, RT:
		label := "LB/LT"
		if button == RT {
			label = "RB/RT"
		}
		if p.family == PlayStation {
			label = "L1/L2"
			if button == RT {
				label = "R1/R2"
			}
		}
		text(dst, label, x, y+3, c)
	case Start:
		for row := 2; row <= 10; row += 4 {
			line(dst, x, y+row, x+12, y+row, c)
		}
	case Back:
		line(dst, x, y+2, x+8, y+2, c)
		line(dst, x, y+2, x, y+8, c)
		line(dst, x+4, y+6, x+12, y+6, c)
		line(dst, x+4, y+6, x+4, y+12, c)
		line(dst, x+4, y+12, x+12, y+12, c)
		line(dst, x+12, y+6, x+12, y+12, c)
	case DPad:
		line(dst, x+6, y, x+6, y+12, c)
		line(dst, x, y+6, x+12, y+6, c)
	}
}

// Overlay composes into separate host pixels. It never mutates a source frame.
// Hidden overlays reuse the source pointer; dst is reusable host storage.
//
// The hint panel is drawn over the bottom of the frame, which keeps its size:
// a frame that grows while hints are visible makes hosts resize or rescale
// the window and push the table around. The top of the frame (the dot-matrix
// display) is never covered.
func Overlay(dst, src *image.RGBA, p *Input, rows []Hint) *image.RGBA {
	if !p.Connected() || len(rows) == 0 || src.Rect.Dx() < 260 || src.Rect.Dy() < 150 {
		return src
	}
	// 640x240 frames are shown with doubled scanlines, so their pixels are
	// twice as tall as wide. Doubling the panel horizontally keeps it square.
	scale := 1
	if src.Rect.Dx() >= 520 {
		scale = 2
	}
	if dst == nil || dst.Rect != src.Rect {
		dst = image.NewRGBA(src.Rect)
	}
	copy(dst.Pix, src.Pix)

	width, height := 300, 20+len(rows)*16
	if max := src.Rect.Dy() / 2; height > max {
		height = max
	}
	panel := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(panel, panel.Rect, image.NewUniform(color.NRGBA{12, 18, 26, 205}), image.Point{}, draw.Src)
	title := "GAMEPAD"
	switch p.family {
	case Xbox:
		title = "XBOX"
	case PlayStation:
		title = "PLAYSTATION"
	case Nintendo:
		title = "NINTENDO"
	}
	text(panel, title, 8, 4, color.RGBA{170, 185, 200, 255})
	for i, row := range rows {
		y := 16 + i*16
		glyph(panel, p, row.Button, 8, y-2)
		text(panel, row.Text, 48, y+1, ink)
	}
	if scale == 2 {
		wide := image.NewRGBA(image.Rect(0, 0, width*2, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				c := panel.RGBAAt(x, y)
				wide.SetRGBA(2*x, y, c)
				wide.SetRGBA(2*x+1, y, c)
			}
		}
		panel = wide
	}
	x := src.Rect.Min.X + (src.Rect.Dx()-panel.Rect.Dx())/2
	y := src.Rect.Max.Y - height - 4
	draw.Draw(dst, image.Rect(x, y, x+panel.Rect.Dx(), y+height), panel, image.Point{}, draw.Over)
	return dst
}
