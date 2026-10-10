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
func Overlay(dst, src *image.RGBA, p *Input, rows []Hint) *image.RGBA {
	if !p.Connected() || len(rows) == 0 || src.Rect.Dx() < 260 || src.Rect.Dy() < 150 {
		return src
	}
	// Preserve the complete source image, with its original scanline aspect.
	// Hints occupy an additional host-only strip below it.
	scale := 1
	if src.Rect.Size() == image.Pt(640, 240) {
		scale = 2
	}
	sourceHeight := src.Rect.Dy() * scale
	stripHeight := 28 + len(rows)*16
	bounds := image.Rect(src.Rect.Min.X, src.Rect.Min.Y, src.Rect.Max.X, src.Rect.Min.Y+sourceHeight+stripHeight*scale)
	if dst == nil || dst.Rect != bounds {
		dst = image.NewRGBA(bounds)
	}
	draw.Draw(dst, dst.Rect, image.NewUniform(color.RGBA{0, 0, 0, 255}), image.Point{}, draw.Src)
	for row := 0; row < src.Rect.Dy(); row++ {
		for repeat := 0; repeat < scale; repeat++ {
			draw.Draw(dst, image.Rect(bounds.Min.X, bounds.Min.Y+row*scale+repeat, bounds.Max.X, bounds.Min.Y+row*scale+repeat+1), src, image.Pt(src.Rect.Min.X, src.Rect.Min.Y+row), draw.Src)
		}
	}
	width := 300
	x := bounds.Min.X + (src.Rect.Dx()/scale-width)/2
	y := bounds.Min.Y + sourceHeight + 4
	panel := image.Rect(x, y, x+width, y+stripHeight-8)
	draw.Draw(dst, panel, image.NewUniform(color.NRGBA{12, 18, 26, 195}), image.Point{}, draw.Over)
	title := "GAMEPAD"
	if p.family == Xbox {
		title = "XBOX"
	}
	if p.family == PlayStation {
		title = "PLAYSTATION"
	}
	if p.family == Nintendo {
		title = "NINTENDO"
	}
	text(dst, title, x+8, y+6, color.RGBA{170, 185, 200, 255})
	for i, row := range rows {
		yy := y + 18 + i*16
		glyph(dst, p, row.Button, x+8, yy-2)
		text(dst, row.Text, x+48, yy+1, ink)
	}
	if scale == 2 {
		// Expand the strip in place, backwards so its source pixels stay intact.
		for row := stripHeight - 1; row >= 0; row-- {
			for col := src.Rect.Dx()/scale - 1; col >= 0; col-- {
				c := dst.RGBAAt(bounds.Min.X+col, bounds.Min.Y+sourceHeight+row)
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						dst.SetRGBA(bounds.Min.X+col*scale+dx, bounds.Min.Y+sourceHeight+row*scale+dy, c)
					}
				}
			}
		}
	}
	return dst
}
