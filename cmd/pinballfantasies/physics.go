package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"pinballfantasies/internal/diagnostics"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/platform"
	"time"
)

func runPhysics(dataDir, output string, duration time.Duration, ticks, releaseAt int) error {
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
	game := physics.New(table)
	completed := 0
	next := func(c platform.PhysicsControls) (*image.RGBA, error) {
		if ticks > 0 && completed >= ticks {
			return game.Frame(), nil
		}
		if completed == releaseAt {
			game.Release(32, 0)
		}
		if c.Release {
			game.Release(game.SpringPosition, 0)
			game.SpringPosition = 0
		} else if c.Down && game.SpringPosition < 32 {
			game.SpringPosition++
		}
		if err := game.Sync(physics.Inputs{Left: c.Left, Right: c.Right, Tilt: c.Tilt}); err != nil {
			return nil, err
		}
		completed++
		if ticks > 0 && completed == ticks {
			diagnostics.Printf("PF3 froze after %d syncs at ball=(%d,%d), velocity=(%d,%d)\n", completed, game.Ball.PixelX, game.Ball.PixelY, game.Ball.VX, game.Ball.VY)
		}
		return game.Frame(), nil
	}
	if output != "" {
		for n := 0; n < ticks; n++ {
			if _, err := next(platform.PhysicsControls{}); err != nil {
				return err
			}
		}
		frame := game.Frame()
		f, err := os.Create(output)
		if err != nil {
			return err
		}
		err = png.Encode(f, frame)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		diagnostics.Printf("PF3 syncs=%d ball=(%d,%d) velocity=(%d,%d) lost=%t\n", completed, game.Ball.PixelX, game.Ball.PixelY, game.Ball.VX, game.Ball.VY, game.Ball.Lost)
		return closeErr
	}
	return platform.ShowPhysics(game.Frame(), duration, next)
}
