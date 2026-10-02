package speeddevils

import (
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/tablelogic"
	"strings"
)

type music = tablelogic.MusicClock

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
func (g *Game) Cue(label string) { g.music.Priority = 0; g.playJingle(label) }
func (g *Game) playJingle(label string) bool {
	label = strings.ToUpper(label)
	s := g.Display.Jingle(label)
	if g.MusicOff && (label == "S_MAIN" || label == "S_SPRING") {
		s.Position = 55
	}
	accepted := g.music.Play(s, 55, programs.Cues)
	if !accepted {
		g.emit("AudioRejected", label, uint64(s.Priority))
		return false
	}
	g.muteMusicReturn()
	if g.Playback != nil {
		g.Playback.Force(int(s.Position))
	}
	g.emit("Music", label, uint64(s.Position))
	return true
}

// Background orders are 0..9 (spring and every main-music segment).
// Muting also applies to saved returns from jingles and rule-script overrides.
func (g *Game) muteMusicReturn() {
	if g.MusicOff && g.music.ReturnPosition <= 9 {
		g.music.ReturnPosition = 55
	}
}
func (g *Game) audioTick() {
	g.muteMusicReturn()
	if g.Playback != nil {
		g.AudioPCM = g.Playback.Sync()
	}
	g.music.Sync(programs.Cues, func(kind string, value uint64) {
		label := "JINGLE_HANDLER"
		if kind == "AudioComplete" {
			label = "JINGLEREADY"
		}
		g.emit(kind, label, value)
	})
}

var sounds = map[string]audio.Effect{
	"SBRICKNER": {22, 18}, "SBRICKUPP": {23, 18}, "SBUMPER": {24, 18}, "SFLIPPUPP": {25, 22}, "SRINNER": {26, 18}, "SNEWBALL": {28, 18}, "SKICKER": {29, 18}, "SFJADER": {30, 18}, "SGROP": {23, 18}, "SBYGEL2": {27, 12}, "S_TOUCH1": {27, 14}, "S_SCORELJUD": {27, 18}, "S_MULTILJUD": {27, 10},
}

func (g *Game) sound(label string) {
	g.emit("Sound", label, 0)
	if g.Playback != nil {
		if s, ok := sounds[label]; ok {
			g.Playback.SoundEffect(s)
		}
	}
}
