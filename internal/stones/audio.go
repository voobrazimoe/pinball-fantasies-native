package stones

import (
	"pinballfantasies/internal/audio"
	"strings"
)

func (g *Game) AttachAudio(m *audio.Module) {
	g.Playback = audio.New(m)
	if g.music.Active {
		g.Playback.Force(int(g.music.Position))
	}
	g.Playback.Jump = func(next int) int {
		if g.music.JumpCount == 1 {
			return int(g.music.ReturnPosition)
		}
		return next
	}
}
func (g *Game) Cue(l string) {
	g.music.Priority = 0
	g.playJingle(l)
	if strings.ToUpper(l) == "S_NOHIGH" {
		g.music.ReturnPosition = 9
	}
}
func (g *Game) playJingle(l string) bool {
	l = strings.ToUpper(l)
	s := g.Display.Jingle(l)
	if g.MusicOff && (l == "S_MAIN" || l == "S_SPRING" || l == "S_SPRING1") {
		s.Position = 52
	}
	if !g.music.Play(s, 52, programs.Cues) {
		return false
	}
	g.muteReturn()
	if g.Playback != nil {
		g.Playback.Force(int(s.Position))
	}
	g.emit("Music", l, uint64(s.Position))
	return true
}
func (g *Game) muteReturn() {
	if g.MusicOff && g.music.ReturnPosition <= 9 {
		g.music.ReturnPosition = 52
	}
}
func (g *Game) audioTick() {
	g.muteReturn()
	if g.Playback != nil {
		g.AudioPCM = g.Playback.Sync()
	}
	g.music.Sync(programs.Cues, func(k string, n uint64) { g.emit(k, "JINGLE_HANDLER", n) })
}
func (g *Game) soundSpec(l string) audio.Effect {
	b := g.Display.Content.Texts[l]
	return audio.Effect{Sample: int(b[0]), Note: int(b[1])}
}
func (g *Game) sound(l string) {
	g.emit("Sound", l, 0)
	if g.Playback != nil {
		if l == "S_TOUCH1" {
			g.Playback.SoundEffectVolume(g.soundSpec(l), 40)
		} else {
			g.Playback.SoundEffect(g.soundSpec(l))
		}
	}
}
func (g *Game) ToggleMusic() {
	g.MusicOff = !g.MusicOff
	if g.MusicOff {
		bg := g.music.Position <= 9
		g.music.ReturnPosition = 52
		if bg {
			g.task(func() bool { g.Cue("S_EMPTY"); return true })
		}
	} else {
		g.music.ReturnPosition = 1
		if g.inChute {
			g.music.ReturnPosition = 0
		}
		if g.music.JumpCount < 1 {
			g.music.JumpCount = 1
		}
	}
}
