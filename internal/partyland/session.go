package partyland

import (
	"image"
	"pinballfantasies/internal/presentation"
)

// SetHighScore installs immutable frontend content. Direct PF4 development keeps
// the pre-PF6 branch disabled unless explicitly supplied, preserving its oracle.
func (g *Game) SetHighScore(top Decimal) { g.highScore = &top }
func (g *Game) checkHighScore() bool {
	if g.inChute || g.special() {
		return false
	}
	return g.beatHighScore()
}

// PLAND/FANTASIE _BEATEN_MATRIX omits CHECKHIGHSCORE's chute/mode guards.
func (g *Game) beatHighScore() bool {
	if g.highScore == nil || g.alreadyBeaten || g.Score.Uint64() <= g.highScore.Uint64() {
		return false
	}
	g.alreadyBeaten = true
	g.emit("HighScoreBeaten", "CHECKHIGHSCORE", g.Score.Uint64())
	return true
}
func (g *Game) Result() (Decimal, bool) { return g.Score, g.Phase == GameOver }
func (g *Game) InChute() bool           { return g.Physics.SpringValid }
func (g *Game) PCM() []byte             { return g.AudioPCM }
func (g *Game) Cue(label string)        { g.Audio.Priority = 0; g.playJingle(label) }

// PresentationAudioSync advances only the native music clock in attract/entry;
// never physics, rule tasks or gameplay Tick. Paused sessions never call it.
func (g *Game) PresentationAudioSync() { g.audioTick() }

// AttractFrame follows VBLANK_INT_DEMO's one-line triangular table scroll,
// without mutating the suspended session's raster, ball, lamps or rule clocks.
func (g *Game) AttractFrame(ticks int) *image.RGBA {
	out := g.Frame()
	field := g.Physics.Table.Initial.Playfield
	var lamps [57]bool
	palette := g.paletteFor(lamps)
	presentation.AttractLamps(ticks, g.Display.Content.LampFlash, func(n int, on bool) { g.applyLamp(&palette, n, on) })
	height := g.Physics.Settings.FieldHeight()
	limit := 576 - height
	y := limit - ticks%(2*limit)
	if y < 0 {
		y = -y
	}
	if g.Physics.Settings.TableY() != 0 {
		y, height = 0, 576
	}
	for row := 0; row < height; row++ {
		for x := 0; x < 320; x++ {
			index := int(field.Indices[(y+row)*320+x]) * 3
			p := (row+g.Physics.Settings.TableY())*out.Stride + x*4
			copy(out.Pix[p:p+3], palette[index:index+3])
		}
	}
	return out
}
