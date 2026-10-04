package gameshow

import (
	"image"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/presentation"
	"strconv"
	"strings"
)

func (g *Game) lampRecord(n int) lamp {
	c := programs.Lamps[strconv.Itoa(n)]
	b := g.Display.Record(c.Ref)
	if len(b) < 2 || len(b) != 2+int(b[1])*3 {
		panic("invalid PRG lamp record")
	}
	c.Start = int(b[0])
	c.RGB = b[2:]
	return c
}
func (g *Game) applyLamp(p *[768]byte, n int, on bool) {
	c := g.lampRecord(n)
	for i, v := range g.Physics.LampRGB(c.RGB, on) {
		p[3*c.Start+i] = v
	}
}
func (g *Game) basePalette() [768]byte {
	p := g.Physics.BasePalette(assets.VGAPalette(g.Physics.Table.Initial.Playfield.Palette))
	for n := 1; n <= 38; n++ {
		g.applyLamp(&p, n, false)
	}
	return p
}
func (g *Game) Palette() [768]byte { return g.palette }
func (g *Game) Frame() *image.RGBA {
	p := presentation.MatrixPaletteMode(g.palette, g.Physics.ReferenceMode, 153)
	d := *g.Display
	return presentation.ComposeNative(g.Physics.FramePalette(p), &d, p, 114, 153, g.Physics.PresentationSettings(), g.Physics.ScreenOffset)
}
func (g *Game) AttractFrame(tick int) *image.RGBA {
	out := g.Frame()
	p := g.basePalette()
	presentation.AttractLamps(tick, g.Display.Content.LampFlash, func(n int, on bool) { g.applyLamp(&p, n, on) })
	bottom := int(g.Physics.BottomRaster())
	y := bottom - tick%(2*bottom)
	if y < 0 {
		y = -y
	}
	field := g.Physics.Table.Initial.Playfield
	if g.Physics.PresentationSettings().TableY() != 0 {
		y = 0
	}
	for row := 0; row < g.Physics.PresentationSettings().RenderHeight(); row++ {
		for x := 0; x < 320; x++ {
			i := int(field.Indices[(y+row)*320+x]) * 3
			o := (row+g.Physics.PresentationSettings().TableY())*out.Stride + x*4
			copy(out.Pix[o:o+3], p[i:i+3])
		}
	}
	return out
}
func (g *Game) matrixNumber(l string) string {
	switch l {
	case "SIFFRORNA":
		return g.Score.String()
	case "BONUSSIFFRORNA":
		return g.Bonus.String()
	case "JACKVALUE":
		return g.Jackpot.String()
	case "CASHPOTVAL":
		return g.CashPot.String()
	case "CASHPOT5VAL":
		return g.CashPot5.String()
	case "RM":
		return g.RaisingMillions.String()
	case "CYCLONECOUNTERBCD", "SKILLCOUNTER":
		return strconv.Itoa(int(g.Skills))
	case "CYCLONESCOREBCD":
		return strconv.FormatUint(uint64(g.Skills)*100000, 10)
	case "TM_TOTAL":
		return g.MoneyTotal.String()
	case "SPINSCOREPTR":
		return strconv.FormatUint(g.spinScore, 10)
	}
	b := g.Display.Content.Texts[l]
	if len(b) == 12 {
		var s strings.Builder
		for _, v := range b {
			s.WriteByte('0' + v)
		}
		return strings.TrimLeft(s.String(), "0")
	}
	return ""
}
