package physics

import (
	"encoding/binary"
	"image"
)

// Frame applies original FLIPPRA delta records before PUTTHEBALL. Each record
// copies original four-plane graphics; collision bits never substitute for art.
func (g *Game) Frame() *image.RGBA {
	for i := range g.Flippers {
		g.animateFlipper(i)
	}
	var foreground []byte
	if g.Ball.High {
		foreground = g.Table.upperForeground
	}
	viewport := int(g.Raster>>4) - 33
	if g.Settings.TableY() != 0 {
		viewport = 0
	}
	return g.Table.Initial.SimulationFrameHeight(g.indices, image.Pt(int(g.Ball.PixelX), int(g.Ball.PixelY)), viewport, foreground, g.Settings.RenderHeight())
}

func (g *Game) animateFlipper(i int) {
	f := g.Flippers[i]
	old := g.graphicsFrame[i]
	now := f.Frame
	if old == now {
		return
	}
	g.graphicsFrame[i] = now
	base := g.Table.deltaBase[i]
	size := g.Table.deltaSize[i]
	distance := int(now - old)
	record := base + int(old)*size
	if distance < 0 {
		distance = -distance
		record = g.Table.deltaMax[i] - int(old)*size
	}
	for distance > 0 {
		chunk := distance
		if chunk > 9 {
			chunk = 9
		}
		for countDown := chunk; countDown > 0; countDown-- {
			count := int(binary.LittleEndian.Uint16(g.Table.deltas[record+2*(countDown-1):]))
			for n := 0; n < count; n++ {
				o := record + 18 + 4*n
				dst := int(binary.LittleEndian.Uint16(g.Table.deltas[o:]))
				src := int(binary.LittleEndian.Uint16(g.Table.deltas[o+2:])) - 0xd4f4
				y, q := dst/84, dst%84
				for plane := 0; plane < 4; plane++ {
					x := int(f.Left) + 4*q + plane
					yy := int(f.Top) + y
					if x < 320 && yy < 576 && src >= 0 && src < g.Table.flipStride {
						g.indices[yy*320+x] = g.Table.flipGraphics[plane*g.Table.flipStride+src]
					}
				}
			}
			record += size
		}
		distance -= chunk
	}
}
