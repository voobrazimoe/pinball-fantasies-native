package gameshow

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
func (g *Game) Cue(label string) {
	g.music.Priority = 0
	g.playJingle(label)
	if strings.ToUpper(label) == "S_NOHIGH" {
		g.music.ReturnPosition = 9
	}
}
func (g *Game) playJingle(label string) bool {
	label = strings.ToUpper(label)
	s := g.Display.Jingle(label)
	if g.MusicOff && (label == "S_MAIN" || label == "S_SPRING") {
		s.Position = 55
	}
	if !g.music.Play(s, 55, programs.Cues) {
		return false
	}
	g.muteMusicReturn()
	if g.Playback != nil {
		g.Playback.Force(int(s.Position))
	}
	g.emit("Music", label, uint64(s.Position))
	return true
}
func (g *Game) muteMusicReturn() {
	if g.MusicOff && g.music.ReturnPosition <= 6 {
		g.music.ReturnPosition = 55
	}
}
func (g *Game) audioTick() {
	g.muteMusicReturn()
	if g.Playback != nil {
		g.AudioPCM = g.Playback.Sync()
	}
	g.music.Sync(programs.Cues, func(k string, n uint64) { g.emit(k, "JINGLE_HANDLER", n) })
}
func (g *Game) soundSpec(label string) audio.Effect {
	b := g.Display.Content.Texts[label]
	return audio.Effect{Sample: int(b[0]), Note: int(b[1])}
}
func (g *Game) sound(label string) {
	g.emit("Sound", label, 0)
	if g.Playback != nil {
		g.Playback.SoundEffect(g.soundSpec(label))
	}
}
func (g *Game) ToggleMusic() {
	g.MusicOff = !g.MusicOff
	if g.MusicOff {
		bg := g.music.Position <= 6
		g.music.ReturnPosition = 55
		if bg {
			g.task(func() bool { g.Cue("S_EMPTY"); return true })
		}
		return
	}
	g.music.ReturnPosition = 1
	if g.inChute {
		g.music.ReturnPosition = 0
	}
	if g.music.JumpCount < 1 {
		g.music.JumpCount = 1
	}
}

func (g *Game) soundVolume(label string, volume uint8) {
	g.emit("Sound", label, uint64(volume))
	if g.Playback != nil {
		g.Playback.SoundEffectVolume(g.soundSpec(label), volume)
	}
}
