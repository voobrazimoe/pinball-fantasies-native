package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/platform"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"time"
)

func runSpeedDevils(dataDir, output string, duration time.Duration, ticks, releaseAt int) error {
	return runNativeTable(2, dataDir, output, duration, ticks, releaseAt)
}
func runGameshow(dataDir, output string, duration time.Duration, ticks, releaseAt int) error {
	return runNativeTable(3, dataDir, output, duration, ticks, releaseAt)
}
func runNativeTable(table int, dataDir, output string, duration time.Duration, ticks, releaseAt int) error {
	if ticks < 0 || releaseAt < -1 {
		return fmt.Errorf("ticks must be nonnegative; release-at must be -1 or nonnegative")
	}
	data, e := os.ReadFile(filepath.Join(dataDir, fmt.Sprintf("TABLE%d.PRG", table)))
	if e != nil {
		return e
	}
	decode := physics.DecodeSpeedDevils
	name := "Speed Devils"
	milestone := 7
	if table == 3 {
		decode = physics.DecodeGameshow
		name = "Billion Dollar Gameshow"
		milestone = 9
	}
	if table == 4 {
		decode = physics.DecodeStones
		name = "Stones N Bones"
		milestone = 10
	}
	t, e := decode(data)
	if e != nil {
		return e
	}
	var g interface {
		frontend.Session
		ToggleMusic()
		SetOriginalBallSetting(byte)
	}
	var summary func() string
	if table == 2 {
		v := speeddevils.New(t, data)
		g = v
		summary = func() string {
			return fmt.Sprintf("score=%s bonus=%s multiplier=%d ball=%d phase=%d", v.Score, v.Bonus, v.Multiplier, v.BallNumber, v.Phase)
		}
	} else if table == 4 {
		v := stones.New(t, data)
		g = v
		summary = func() string {
			return fmt.Sprintf("score=%s bonus=%s multiplier=%d ball=%d phase=%d", v.Score, v.Bonus, v.Multiplier, v.BallNumber, v.Phase)
		}
	} else {
		v := gameshow.New(t, data)
		g = v
		summary = func() string {
			return fmt.Sprintf("score=%s bonus=%s multiplier=%d ball=%d phase=%d", v.Score, v.Bonus, v.Multiplier, v.BallNumber, v.Phase)
		}
	}
	b, e := os.ReadFile(filepath.Join(dataDir, fmt.Sprintf("TABLE%d.MOD", table)))
	if e != nil {
		return e
	}
	decodeMOD := audio.DecodeSpeedDevils
	if table == 3 {
		decodeMOD = audio.DecodeGameshow
	}
	if table == 4 {
		decodeMOD = audio.DecodeStones
	}
	m, e := decodeMOD(b)
	if e != nil {
		return e
	}
	g.(interface{ AttachAudio(*audio.Module) }).AttachAudio(m)
	if config, e := os.ReadFile(filepath.Join(dataDir, "PINBALL.CFG")); e == nil && len(config) == 6 {
		g.SetOriginalBallSetting(config[0])
	}
	var device *platform.AudioDevice
	if output == "" {
		device, e = platform.OpenAudio()
		if e != nil {
			if fatal := platform.AudioStartupError(e); fatal != nil {
				return fatal
			}
			fmt.Fprintln(os.Stderr, "Audio output unavailable:", e)
		} else {
			defer device.Close()
		}
	}
	completed := 0
	next := func(c platform.PhysicsControls) (*image.RGBA, error) {
		if ticks > 0 && completed >= ticks {
			return g.Frame(), nil
		}
		if c.MusicToggle {
			g.ToggleMusic()
		}
		if completed == releaseAt {
			g.Release(32, 0)
		}
		if e := g.Sync(physics.Inputs{Left: c.Left, Right: c.Right, Down: c.Down, Release: c.Release, Tilt: c.Tilt}); e != nil {
			return nil, e
		}
		completed++
		if device != nil {
			if e := device.Queue(g.PCM()); e != nil {
				return nil, e
			}
		}
		return g.Frame(), nil
	}
	if output != "" {
		for completed < ticks {
			if _, e := next(platform.PhysicsControls{}); e != nil {
				return e
			}
		}
		f, e := os.Create(output)
		if e != nil {
			return e
		}
		e = png.Encode(f, g.Frame())
		ce := f.Close()
		if e != nil {
			return e
		}
		fmt.Printf("PF%d ticks=%d %s\n", milestone, completed, summary())
		return ce
	}
	return platform.ShowTablePhysics(g.Frame(), duration, name, next)
}
