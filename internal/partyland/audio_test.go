package partyland

import (
	"bytes"
	"os"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"testing"
)

func audioModule(t *testing.T) *audio.Module {
	t.Helper()
	testinputs.Require(t, "../../TABLE1.MOD")
	b, e := os.ReadFile("../../TABLE1.MOD")
	if e != nil {
		t.Fatal(e)
	}
	m, e := audio.Decode(b)
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func TestPF5AudioGameplayCoupling(t *testing.T) {
	actions := map[string]func(*Game){
		"Mystery":       func(g *Game) { g.Arcade = true; g.arcade() },
		"CrazyLetter":   func(g *Game) { g.crazy() },
		"ExtraBall":     func(g *Game) { g.BallFeature = true; g.dragon() },
		"JackpotRepeat": func(g *Game) { g.JackpotNormal = true; g.dragon() },
		"DrainNewBall": func(g *Game) {
			g.Score = Number(1000)
			g.ScoreChanged = true
			g.Bonus = Number(1000)
			g.Multiplier = 2
			g.Cyclones = 2
			g.drain()
		},
		"ForcedReturn": func(g *Game) { g.music("S_MAIN") },
	}
	for prize := 1; prize <= 4; prize++ {
		n := prize
		actions[[]string{"", "ArcadeCrazy", "Arcade5M", "Arcade1M", "Arcade500K"}[prize]] = func(g *Game) { g.spinReward(n) }
	}
	for name, action := range actions {
		t.Run(name, func(t *testing.T) {
			silent, native := newTestGame(t), newTestGame(t)
			native.AttachAudio(audioModule(t))
			quiet(silent)
			quiet(native)
			action(silent)
			action(native)
			var rendered []byte
			for i := 0; i < 1600; i++ {
				if name == "ForcedReturn" && i == 546 {
					silent.music("SJINGLE22")
					native.music("SJINGLE22")
					if native.Audio.ReturnPosition != 2 {
						t.Fatal("saved current order")
					}
				}
				quiet(silent)
				quiet(native)
				in := physics.Inputs{}
				if e := silent.Sync(in); e != nil {
					t.Fatal(e)
				}
				if e := native.Sync(in); e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(silent.Events, native.Events) || silent.Audio != native.Audio || silent.Score != native.Score || silent.Phase != native.Phase || silent.BallNumber != native.BallNumber || !reflect.DeepEqual(silent.matrix, native.matrix) {
					t.Fatalf("native playback changed gameplay at sync%d", i)
				}
				if native.Audio.Active && native.Playback.Order != int(native.Audio.Position) {
					t.Fatalf("audible order%d differs from semantic order%d at sync%d", native.Playback.Order, native.Audio.Position, i)
				}
				if i%11 == 0 {
					rendered = append(rendered, native.AudioPCM...)
				} // discarded host buffers cannot influence state.
			}
			if bytes.Equal(rendered, make([]byte, len(rendered))) {
				t.Fatal("silent renderer")
			}
		})
	}
}
func TestPF5PriorityAndReadyState(t *testing.T) {
	g := newTestGame(t)
	quiet(g)
	g.AttachAudio(audioModule(t))
	g.music("SJINGLE4")
	if g.Audio.ReadyAnim || g.Audio.ReadyLogic {
		t.Fatal("accepted jingle didn't clear readiness")
	}
	before := g.Playback.Order
	g.music("SJINGLE22")
	if g.Playback.Order != before || g.Audio.JumpCount != 1 {
		t.Fatal("rejected priority or ASM swap semantics")
	}
	ticks(t, g, 500)
	if !g.Audio.ReadyAnim || !g.Audio.ReadyLogic || g.Audio.Priority != g.Audio.LastPriority {
		t.Fatal("completion restore")
	}
	g.sound("SBUMPER1")
	state := g.Audio
	ticks(t, g, 1)
	if g.Audio.ReadyAnim != state.ReadyAnim || g.Audio.ReadyLogic != state.ReadyLogic {
		t.Fatal("PCM fabricated readiness")
	}
}
func TestPF5OfflineGameplayDeterminism(t *testing.T) {
	a, b := newTestGame(t), newTestGame(t)
	m := audioModule(t)
	a.AttachAudio(m)
	b.AttachAudio(m)
	for i := 0; i < 1200; i++ {
		if i == 100 || i == 700 {
			a.Release(32, 0)
			b.Release(32, 0)
		}
		in := physics.Inputs{Left: i%93 < 18, Right: i%71 < 14}
		if e := a.Sync(in); e != nil {
			t.Fatal(e)
		}
		if e := b.Sync(in); e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(a.AudioPCM, b.AudioPCM) || !reflect.DeepEqual(a.Events, b.Events) {
			t.Fatalf("offline divergence at%d", i)
		}
	}
	if a.Score.Uint64() != 2300000 || a.BallNumber != 2 {
		t.Fatal("PF4 script regression", a.Score, a.BallNumber)
	}
}

func TestMusicOffSurvivesJingleReturns(t *testing.T) {
	for _, native := range []bool{false, true} {
		g := newTestGame(t)
		if native {
			g.AttachAudio(audioModule(t))
		}
		g.Cue("S_MAIN")
		g.ToggleMusic()
		g.runTasks()
		for _, label := range []string{"SJINGLE3", "SJINGLE22", "SJINGLE7"} {
			g.Cue(label)
			if g.Audio.ReturnPosition != 62 {
				t.Fatal("muted jingle saved background return", label, g.Audio)
			}
			completed := false
			for tick := 0; tick < 5000; tick++ {
				g.audioTick()
				if g.Audio.JumpCount == 0 {
					completed = true
					break
				}
			}
			if !completed || !g.MusicOff || g.Audio.Position != 62 || (native && g.Playback.Order != 62) {
				t.Fatal("jingle reenabled music", label, g.Audio)
			}
			g.audioTick()
			if native && !bytes.Equal(g.AudioPCM, make([]byte, len(g.AudioPCM))) {
				t.Fatal("background PCM audible after muted jingle", label)
			}
		}
		// Rule scripts can overwrite the saved return after accepting a jingle.
		g.Cue("SJINGLE3")
		before := g.Audio.Position
		g.Audio.ReturnPosition = 3
		g.audioTick()
		if g.Audio.ReturnPosition != 62 || g.Audio.Position != before {
			t.Fatal("muted return override interrupted jingle")
		}
		g.ToggleMusic()
		if g.MusicOff || g.Audio.ReturnPosition != 0 {
			t.Fatal("M did not reenable spring music")
		}
		g.ToggleMusic()
		g.runTasks()
		if g.Audio.Position != before || g.Audio.ReturnPosition != 62 {
			t.Fatal("M interrupted active jingle")
		}
	}
}
