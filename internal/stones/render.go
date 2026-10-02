package stones

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
	for _, n := range programs.LampOrder {
		g.applyLamp(&p, n, false)
	}
	return p
}
func (g *Game) Palette() [768]byte { return g.palette }
func (g *Game) Frame() *image.RGBA {
	p := g.palette
	d := *g.Display
	if !g.matrix.active {
		d.Clear()
		d.Text("PLAYER "+strconv.Itoa(g.Session.CurrentPlayer), 8, 1, 5)
		d.Text("BALL "+strconv.Itoa(int(g.BallNumber)), 8, 9, 5)
		d.Score(g.Score.String())
	}
	return presentation.ComposeNative(g.Physics.FramePalette(p), &d, p, 231, 79, g.Physics.Settings, g.Physics.ScreenOffset)
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
	if g.Physics.Settings.TableY() != 0 {
		y = 0
	}
	for row := 0; row < g.Physics.Settings.RenderHeight(); row++ {
		for x := 0; x < 320; x++ {
			i := int(field.Indices[(y+row)*320+x]) * 3
			o := (row+g.Physics.Settings.TableY())*out.Stride + x*4
			copy(out.Pix[o:o+3], p[i:i+3])
		}
	}
	return out
}
func (g *Game) matrixNumber(l string) string {
	values := map[string]Decimal{"SIFFRORNA": g.Score, "BONUSSIFFRORNA": g.Bonus, "JACKVALUE": g.Jackpot, "TOWERVALUE": g.TowerValue, "VAULTVALUE": g.VaultValue, "WELLVALUE": g.WellValue, "SKILLSCOREBCD": g.SkillScore, "SULP_SCORE": g.SulpScore, "OR_TOTAL": g.GhostTotal, "TM_TOTAL": g.GrimTotal, "CYCLONESCOREBCD": number(uint64(g.Screams) * 100000), "CYCLONECOUNTERBCD": number(uint64(g.Screams))}
	if v, ok := values[l]; ok {
		return v.String()
	}
	b := g.Display.Content.Texts[l]
	if len(b) == 12 {
		var s strings.Builder
		for _, v := range b {
			s.WriteByte('0' + v)
		}
		return s.String()
	}
	return "0"
}

// MATRIXOFF writes three consecutive DAC entries, overlapping MUMMY's lamp.
// Apply writes at command/flash boundaries, before composition, in source order.
func (g *Game) matrixPalette(force bool) {
	if !force && g.matrixOn == g.Display.On {
		return
	}
	g.matrixOn = g.Display.On
	label := "MATRIXOFF"
	if g.matrixOn {
		label = "MATRIXON"
	}
	b := g.Display.Content.Texts[label]
	rgb := b[2 : 2+int(b[1])*3]
	if g.matrixOn {
		rgb = g.Physics.LampRGB(rgb, true)
	} else {
		rgb = append([]byte(nil), rgb...)
		for i, v := range rgb {
			rgb[i] = assets.DACRGB(v)
		}
	}
	for i, v := range rgb {
		g.palette[int(b[0])*3+i] = v
	}
}
