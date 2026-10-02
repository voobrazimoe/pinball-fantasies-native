package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/platform"
	"time"
)

func runRules(dataDir, output string, duration time.Duration, ticks, releaseAt int, scripted bool) error {
	if ticks < 0 || releaseAt < -1 {
		return fmt.Errorf("ticks must be nonnegative; release-at must be -1 or nonnegative")
	}
	data, err := os.ReadFile(filepath.Join(dataDir, "TABLE1.PRG"))
	if err != nil {
		return err
	}
	table, err := physics.DecodePartyLand(data)
	if err != nil {
		return err
	}
	game := partyland.New(table, data)
	moduleData, err := os.ReadFile(filepath.Join(dataDir, "TABLE1.MOD"))
	if err != nil {
		return err
	}
	module, err := audio.Decode(moduleData)
	if err != nil {
		return err
	}
	game.AttachAudio(module)
	var device *platform.AudioDevice
	if output == "" {
		device, err = platform.OpenAudio()
		if err != nil {
			if fatal := platform.AudioStartupError(err); fatal != nil {
				return fatal
			}
			fmt.Fprintln(os.Stderr, "Audio output unavailable; gameplay continues:", err)
		} else {
			defer device.Close()
		}
	}
	if config, e := os.ReadFile(filepath.Join(dataDir, "PINBALL.CFG")); e == nil && len(config) == 6 {
		game.SetOriginalBallSetting(config[0])
	}
	completed := 0
	next := func(c platform.PhysicsControls) (*image.RGBA, error) {
		if ticks > 0 && completed >= ticks {
			return game.Frame(), nil
		}
		if scripted {
			c.Left = completed%93 < 18
			c.Right = completed%71 < 14
			if completed == 100 || completed == 700 {
				game.Release(32, 0)
			}
			c.Release = false
		}
		if c.MusicToggle {
			game.ToggleMusic()
		}
		if completed == releaseAt {
			game.Release(32, 0)
		}
		if err := game.Sync(physics.Inputs{Left: c.Left, Right: c.Right, Down: c.Down, Release: c.Release, Tilt: c.Tilt}); err != nil {
			return nil, err
		}
		if device != nil {
			if err := device.Queue(game.AudioPCM); err != nil {
				return nil, err
			}
		}
		completed++
		return game.Frame(), nil
	}
	if output != "" {
		// Headless scripts need no intermediate frame; flipper deltas compose at end.
		for completed < ticks {
			if completed == releaseAt || (scripted && (completed == 100 || completed == 700)) {
				game.Release(32, 0)
			}
			input := physics.Inputs{}
			if scripted {
				input.Left = completed%93 < 18
				input.Right = completed%71 < 14
			}
			if err := game.Sync(input); err != nil {
				return err
			}
			completed++
		}
		f, err := os.Create(output)
		if err != nil {
			return err
		}
		err = png.Encode(f, game.Frame())
		closeErr := f.Close()
		if err != nil {
			return err
		}
		fmt.Printf("PF4 ticks=%d score=%s bonus=%s multiplier=%d ball=%d phase=%d\n", completed, game.Score, game.Bonus, game.Multiplier, game.BallNumber, game.Phase)
		return closeErr
	}
	return platform.ShowPhysics(game.Frame(), duration, next)
}
