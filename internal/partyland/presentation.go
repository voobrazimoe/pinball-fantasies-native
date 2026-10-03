package partyland

import (
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/physics"
	"strconv"
	"strings"
)

func (g *Game) matrixNumber(label string) string {
	switch label {
	case "SIFFRORNA":
		return g.Score.String()
	case "BONUSSIFFRORNA":
		return g.Bonus.String()
	case "JACKVALUE", "JACKPOT":
		return g.Jackpot.String()
	case "CYCLONECOUNTERBCD":
		return strconv.Itoa(int(g.Cyclones))
	case "CYCLONESCOREBCD":
		return strconv.FormatUint(uint64(g.Cyclones)*100000, 10)
	case "SKILL_SCORE":
		return g.SkillTunnel.String()
	case "SKILL_SCORE2":
		return g.SkillCyclone.String()
	case "HAPPY_HOUR_TOTAL", "OR_TOTAL":
		return g.HappyTotal.String()
	case "MEGA_LAUGH_TOTAL", "TM_TOTAL":
		return g.MegaTotal.String()
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
		g.music("S_DANGER")
	}
	if tilted {
		g.music("S_TILT")
		g.Audio.ReturnPosition = 62
		g.effectEnded = true
		g.flashes = [15]flash{}
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
		background := g.Audio.Position <= 5
		g.Audio.ReturnPosition = 62
		if background {
			g.task(func() bool { g.Cue("S_EMPTY"); return true })
		}
		return
	}
	g.Audio.ReturnPosition = 1
	if g.inChute {
		g.Audio.ReturnPosition = 0
	}
	if g.Audio.JumpCount < 1 {
		g.Audio.JumpCount = 1
	}
}
