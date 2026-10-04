package assets

import (
	"image"
	"image/draw"
	"os"
)

const BallSize = 16
const foregroundOffset = 0x300b0

// InitialState freezes PLAND/SETBALL immediately after SETBALLPOS, before
// KollaKulan can move it. Coordinates are in PF1 playfield pixels; the separate
// SPLH=33 score panel is excluded. The supplied installation selects HI_RES.
type InitialState struct {
	Viewport   image.Rectangle
	BallOrigin image.Point
	BallHigh   bool
}

func PartyLandInitialState() InitialState {
	return InitialState{
		Viewport:   image.Rect(0, 576-(350-33), Width, Height),
		BallOrigin: image.Pt((310-BallSize/2)-5, (543-BallSize/2)-5),
	}
}

// InitialTable holds the shared original indexed playfield/sprite presentation.
type InitialTable struct {
	Playfield *Playfield
	State     InitialState
	// Zero pixels are absent from the unrolled original sprite, not painted.
	Ball       [BallSize * BallSize]byte
	foreground []byte
}

// InitialPartyLand preserves the PF2 API.
type InitialPartyLand = InitialTable

func LoadInitialPartyLand(path string) (*InitialPartyLand, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return DecodeInitialPartyLand(data)
}

func DecodeInitialPartyLand(data []byte) (*InitialPartyLand, error) {
	// The shared layout profile bounds all consumed regions before fixed-offset
	// artwork and physics records are read.
	field, err := DecodePartyLand(data)
	if err != nil {
		return nil, err
	}
	p := &InitialPartyLand{Playfield: field, State: PartyLandInitialState(),
		foreground: append([]byte(nil), data[foregroundOffset:foregroundOffset+Width/8*Height]...)}
	for i, offset := range ballPaletteOffsets {
		if offset != 0 {
			p.Ball[i] = data[offset]
		}
	}
	return p, nil
}

// Framebuffer reproduces PUTTHEBALL's first visible lower-plane draw. Original
// PUTBALL suppresses each sprite pixel when HID1 has its foreground bit set.
// Saved-background restoration is unnecessary for this single frozen frame.
func (p *InitialTable) Framebuffer() *image.RGBA {
	full := p.Playfield.Framebuffer()
	for i, c := range p.Ball {
		if c == 0 {
			continue
		}
		x, y := p.State.BallOrigin.X+i%BallSize, p.State.BallOrigin.Y+i/BallSize
		if !image.Pt(x, y).In(full.Rect) {
			continue
		}
		if p.foreground[y*(Width/8)+x/8]&(0x80>>uint(x&7)) != 0 {
			continue
		}
		o := full.PixOffset(x, y)
		copy(full.Pix[o:o+3], p.Playfield.Palette[int(c)*3:int(c)*3+3])
	}
	dst := image.NewRGBA(image.Rect(0, 0, p.State.Viewport.Dx(), p.State.Viewport.Dy()))
	draw.Draw(dst, dst.Rect, full, p.State.Viewport.Min, draw.Src)
	return dst
}
