package speeddevils

import (
	"bytes"
	"os"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"testing"
)

func TestNativeAudioAndCueClock(t *testing.T) {
	testinputs.Require(t, "../../TABLE2.MOD")
	b, e := os.ReadFile("../../TABLE2.MOD")
	if e != nil {
		t.Fatal(e)
	}
	m, e := audio.DecodeSpeedDevils(b)
	if e != nil {
		t.Fatal(e)
	}
	for _, label := range []string{"S_MAIN", "SJINGLE3", "SJINGLE7", "SJINGLE22", "S_LOSTBALL"} {
		t.Run(label, func(t *testing.T) {
			silent, native := testGame(t), testGame(t)
			native.AttachAudio(m)
			silent.Cue(label)
			native.Cue(label)
			audible := false
			for i := 0; i < 1600; i++ {
				silent.Events = nil
				native.Events = nil
				silent.audioTick()
				native.audioTick()
				if !reflect.DeepEqual(silent.Events, native.Events) || silent.music != native.music {
					t.Fatal("PCM affects semantic clock", i)
				}
				if native.Playback.Order != int(native.music.Position) {
					t.Fatalf("cue order at sync %d: PCM %d logic %d", i, native.Playback.Order, native.music.Position)
				}
				if !bytes.Equal(native.AudioPCM, make([]byte, len(native.AudioPCM))) {
					audible = true
				}
			}
			if !audible {
				t.Fatal("silent table music")
			}
		})
	}
}

func TestMusicOffSurvivesJingleReturns(t *testing.T) {
	testinputs.Require(t, "../../TABLE2.MOD")
	raw, e := os.ReadFile("../../TABLE2.MOD")
	if e != nil {
		t.Fatal(e)
	}
	m, e := audio.DecodeSpeedDevils(raw)
	if e != nil {
		t.Fatal(e)
	}
	for _, native := range []bool{false, true} {
		g := testGame(t)
		if native {
			g.AttachAudio(m)
		}
		g.Cue("S_MAIN")
		g.ToggleMusic()
		g.runTasks()
		for _, label := range []string{"SJINGLE3", "SJINGLE2", "SJINGLE8"} {
			g.Cue(label)
			if g.music.ReturnPosition != 55 {
				t.Fatal("muted jingle saved background return", label, g.music)
			}
			completed := false
			for tick := 0; tick < 5000; tick++ {
				g.audioTick()
				if g.music.JumpCount == 0 {
					completed = true
					break
				}
			}
			if !completed || !g.MusicOff || g.music.Position != 55 || (native && g.Playback.Order != 55) {
				t.Fatal("jingle reenabled music", label, g.music)
			}
			g.audioTick()
			if native && !bytes.Equal(g.AudioPCM, make([]byte, len(g.AudioPCM))) {
				t.Fatal("background PCM audible after muted jingle", label)
			}
		}
		g.Cue("SJINGLE3")
		before := g.music.Position
		g.music.ReturnPosition = 7
		g.audioTick()
		if g.music.ReturnPosition != 55 || g.music.Position != before {
			t.Fatal("muted return override interrupted jingle")
		}
		g.ToggleMusic()
		if g.MusicOff || g.music.ReturnPosition != 0 {
			t.Fatal("M did not reenable spring music")
		}
		g.ToggleMusic()
		g.runTasks()
		if g.music.Position != before || g.music.ReturnPosition != 55 {
			t.Fatal("M interrupted active jingle")
		}
	}
}
