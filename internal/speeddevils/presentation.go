package speeddevils

import (
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/physics"
	"strconv"
	"strings"
)

// The original writes live counter digits into several source text buffers
// before their matrix program prints them. Digits are stored matrix-encoded:
// byte '7'+d, which presentation.SourceText maps back to '0'..'9'; the
// original's blank is '*', which the font has no glyph for.
//
// SDEV.ASM:4795 Put_In_Text
func putInText(buf []byte, at int, value uint16) {
	if at < 0 || at+3 > len(buf) {
		panic("source text write exceeds buffer")
	}
	for i := 2; i >= 0; i-- {
		buf[at+i] = byte(value%10) + '7'
		value /= 10
	}
	// LOOPEN_BERTIL blanks the whole leading run of encoded zeros.
	for i := at; i < at+3 && buf[i] == '7'; i++ {
		buf[i] = '*'
	}
}

// putInTextTopZero mirrors the "lites the jump"/"lites the off road" loops
// (SDEV.ASM:4721, :4755): only the most significant digit is blanked.
func putInTextTopZero(buf []byte, at int, value uint16) {
	if at < 0 || at+3 > len(buf) {
		panic("source text write exceeds buffer")
	}
	for i := 2; i >= 0; i-- {
		d := byte(value % 10)
		if d == 0 && i == 0 {
			buf[at+i] = '*'
		} else {
			buf[at+i] = d + '7'
		}
		value /= 10
	}
}

// sourceTextBuf returns a private mutable copy of a generated text buffer.
// The package-level content map shares byte slices between displays, so a
// write-back must never mutate the extracted data in place.
func (g *Game) sourceTextBuf(label string) []byte {
	return g.Display.MutableText(label, 8)
}

// writeMilesText is NotStandardSeries (SDEV.ASM:4773-4774): "move miles to the
// scrolltext", replacing the XXX of "     XXX MILES      ".
func (g *Game) writeMilesText() {
	{
		buf := g.sourceTextBuf("MILES_TEXT")
		putInText(buf, 5, g.Miles)
	}
}

// writeJumpText is the "lites the jump" block (SDEV.ASM:4721-4729).
func (g *Game) writeJumpText() {
	{
		buf := g.sourceTextBuf("JUMP_AT_TEXT")
		putInTextTopZero(buf, 0, g.NextJump)
	}
}

// writeOffRoadText is the "lites the off road" block (SDEV.ASM:4755-4763).
func (g *Game) writeOffRoadText() {
	{
		buf := g.sourceTextBuf("OFFROAD_AT_TEXT")
		putInTextTopZero(buf, 0, g.NextOffRoad)
	}
}

func (g *Game) matrixNumber(label string) string {
	switch label {
	case "SIFFRORNA":
		return g.Score.String()
	case "BONUSSIFFRORNA":
		return g.Bonus.String()
	case "JACKVALUE", "JACKPOT":
		return g.Jackpot.String()
	case "CYCLONECOUNTERBCD":
		return strconv.Itoa(int(g.Miles))
	case "CYCLONESCOREBCD":
		return strconv.FormatUint(uint64(g.Miles)*100000, 10)
	case "HAPPY_HOUR_TOTAL", "OR_TOTAL":
		return g.OffRoadTotal.String()
	case "MEGA_LAUGH_TOTAL", "TM_TOTAL":
		return g.TurboTotal.String()
	}
	b := g.Display.Content.Texts[label]
	if len(b) == 12 {
		var s strings.Builder
		for _, v := range b {
			s.WriteByte('0' + v)
		}
		return strings.TrimLeft(s.String(), "0")
	}
	return ""
}

func (g *Game) gameplayControl(in physics.Inputs) {
	g.tiltControl(in)
	g.springControl(in)
}
func (g *Game) tiltControl(in physics.Inputs) {
	warning, tilted := g.Physics.TiltInput(in.Tilt, g.inChute)
	if warning {
		g.playJingle("S_DANGER")
	}
	if tilted {
		g.playJingle("S_TILT")
		g.music.ReturnPosition = 62
		g.flashes = [64]flash{}
		for i := 1; i < len(g.Lights); i++ {
			g.light(i, false)
		}
		g.beginMatrix("TILTTS")
		g.emit("Tilt", "HE_TILTED", uint64(g.Physics.TiltCounter))
	}
}

// VBLANK_INT calls SPRINGTASK after DO_ELECTRONICS has run its task slots.
func (g *Game) springControl(in physics.Inputs) {
	gameplay.Spring(&g.Physics.SpringPosition, g.Physics.SpringValid, in, func(charge uint8) {
		g.Release(charge, uint8(g.clock))
	})
}
func (g *Game) presentationTick() { g.matrixTick() }

// MUSIC_TOGGLE mutates only the source main/spring entries; jingles remain active.
func (g *Game) ToggleMusic() {
	g.MusicOff = !g.MusicOff
	if g.MusicOff {
		background := g.music.Position <= 9
		g.music.ReturnPosition = 55
		if background {
			g.task(func() bool { g.Cue("S_EMPTY"); return true })
		}
		return
	}
	g.music.ReturnPosition = 2
	if g.inChute {
		g.music.ReturnPosition = 0
	}
	if g.music.JumpCount < 1 {
		g.music.JumpCount = 1
	}
}
