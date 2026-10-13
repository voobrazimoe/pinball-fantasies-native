package assets

import (
	"encoding/binary"
	"image"
)

// PaletteLUT packs a VGA palette as little-endian RGBA words, alpha 255.
type PaletteLUT [256]uint32

func (l *PaletteLUT) Load(p *[768]byte) {
	for i := range l {
		l[i] = uint32(p[3*i]) | uint32(p[3*i+1])<<8 | uint32(p[3*i+2])<<16 | 0xff000000
	}
}

// RenderRows writes table rows base..base+rows-1 to dst rows dstY.. without
// allocating. It produces the same pixels as Framebuffer followed by the
// viewport copy: rows outside the table get outside instead of artwork.
func RenderRows(dst *image.RGBA, dstY int, indices []byte, lut *PaletteLUT, base, rows int, outside uint32) {
	for k := 0; k < rows; k++ {
		out := dst.Pix[(dstY+k)*dst.Stride:][:Width*4]
		t := base + k
		if t < 0 || t >= Height {
			for x := 0; x < Width; x++ {
				binary.LittleEndian.PutUint32(out[4*x:], outside)
			}
			continue
		}
		src := indices[t*Width:][:Width]
		for x, c := range src {
			binary.LittleEndian.PutUint32(out[4*x:], lut[c])
		}
	}
}

// DrawBall is PUTTHEBALL for RenderRows output: zero sprite pixels are absent
// and HID1 foreground bits hide the ball. upperForeground replaces the lower
// plane's mask when the ball is on the upper level.
func (p *InitialTable) DrawBall(dst *image.RGBA, dstY int, lut *PaletteLUT, origin image.Point, upperForeground []byte, base, rows int) {
	foreground := p.foreground
	if upperForeground != nil {
		foreground = upperForeground
	}
	for i, c := range p.Ball {
		if c == 0 {
			continue
		}
		x, y := origin.X+i%BallSize, origin.Y+i/BallSize
		if x < 0 || x >= Width || y < 0 || y >= Height || y < base || y >= base+rows {
			continue
		}
		if foreground[y*(Width/8)+x/8]&(0x80>>uint(x&7)) != 0 {
			continue
		}
		binary.LittleEndian.PutUint32(dst.Pix[(dstY+y-base)*dst.Stride+4*x:], lut[c])
	}
}
