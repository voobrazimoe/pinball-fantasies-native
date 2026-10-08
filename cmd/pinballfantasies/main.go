package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/platform"
	"pinballfantasies/internal/settings"
)

func run() error {
	data := flag.String("data-dir", platform.DefaultDataDir(), "original installation directory")
	output := flag.String("png", "", "write deterministic PNG and exit without opening a window")
	duration := flag.Duration("duration", 0, "automatically close window after this interval (smoke test)")
	pf2 := flag.Bool("pf2", false, "show/export the unchanged PF2 initial Party Land frame")
	pf1 := flag.Bool("pf1", false, "show/export the unchanged full 320x576 PF1 playfield baseline")
	pf4Script := flag.Bool("pf4-script", false, "PF4 deterministic validation inputs: two launches and timed Shift presses")
	pf10 := flag.Bool("pf10", false, "play native Stones N Bones; Down spring, Space push, Shift/Ctrl/Alt flippers")
	pf9 := flag.Bool("pf9", false, "play native Billion Dollar Gameshow; Down spring, Space push, Shift/Ctrl/Alt flippers")
	pf7 := flag.Bool("pf7", false, "play native Speed Devils; Down spring, Space push, Shift/Ctrl/Alt flippers")
	pf4 := flag.Bool("pf4", false, "play Party Land rules and scoring; Down spring, Space push, Shift/Ctrl/Alt flippers")
	pf3 := flag.Bool("pf3", false, "run original Party Land ball physics; Down spring, Space push, Shift/Ctrl/Alt flippers")
	ticks := flag.Int("ticks", 0, "gameplay deterministic sync count; live window freezes after this many (0 is unlimited)")
	releaseAt := flag.Int("release-at", -1, "gameplay automatic full-charge release before this sync (-1 uses Down Arrow)")
	configDir := flag.String("config-dir", "", "native options directory (default userdata beside executable/AppImage)")
	highDir := flag.String("high-score-dir", "", "native high-score directory (default userdata beside executable/AppImage)")
	flag.Parse()
	if err := platform.InitDiagnostics(*configDir); err != nil {
		return err
	}
	prepare := platform.PrepareDataDir
	if *output == "" && !*pf1 && !*pf2 && !*pf3 && !*pf4 && !*pf7 && !*pf9 && !*pf10 {
		// The live frontend offers importing the full game or the bundled demo.
		prepare = platform.ChooseDataDir
	}
	resolvedData, cleanupData, e := prepare(*data)
	if e != nil {
		return e
	}
	defer cleanupData()
	*data = resolvedData
	if *pf4Script && !*pf4 {
		return fmt.Errorf("-pf4-script requires -pf4")
	}
	if *pf10 {
		if *pf1 || *pf2 || *pf3 || *pf4 || *pf7 || *pf9 || *pf4Script {
			return fmt.Errorf("-pf10 cannot be combined with another development mode")
		}
		return runNativeTable(4, *data, *output, *duration, *ticks, *releaseAt)
	}
	if *pf9 {
		if *pf1 || *pf2 || *pf3 || *pf4 || *pf7 || *pf4Script {
			return fmt.Errorf("-pf9 cannot be combined with another development mode")
		}
		return runGameshow(*data, *output, *duration, *ticks, *releaseAt)
	}
	if *pf7 {
		if *pf1 || *pf2 || *pf3 || *pf4 || *pf4Script {
			return fmt.Errorf("-pf7 cannot be combined with another development mode")
		}
		return runSpeedDevils(*data, *output, *duration, *ticks, *releaseAt)
	}
	if *pf4 {
		if *pf1 || *pf2 || *pf3 {
			return fmt.Errorf("choose one of -pf1, -pf2, -pf3, -pf4")
		}
		return runRules(*data, *output, *duration, *ticks, *releaseAt, *pf4Script)
	}
	if *pf3 {
		if *pf1 || *pf2 {
			return fmt.Errorf("choose one of -pf1, -pf2, -pf3")
		}
		return runPhysics(*data, *output, *duration, *ticks, *releaseAt)
	}
	if *pf1 && *pf2 {
		return fmt.Errorf("choose either -pf1 or -pf2")
	}
	if *ticks < 0 {
		return fmt.Errorf("ticks must be nonnegative")
	}
	if !*pf1 && !*pf2 {
		directory, e := platform.StateDirectory(*highDir)
		if e != nil {
			return e
		}
		optionsDir, e := platform.StateDirectory(*configDir)
		if e != nil {
			return e
		}
		for _, state := range []string{directory, optionsDir} {
			if e := platform.CheckStateSeparate(*data, state); e != nil {
				return e
			}
		}
		r, e := frontend.LoadConfigured(*data, frontend.FileStore{Directory: directory, SeedDirectory: *data}, &settings.Store{Directory: optionsDir, SeedDirectory: *data})
		if e != nil {
			return e
		}
		if *output != "" {
			for i := 0; i < *ticks; i++ {
				if e := r.Update(frontend.Input{}); e != nil {
					return e
				}
			}
			f, e := os.Create(*output)
			if e != nil {
				return e
			}
			e = png.Encode(f, r.Frame())
			ce := f.Close()
			if e != nil {
				return e
			}
			return ce
		}
		device, e := platform.OpenAudio()
		if e != nil {
			if fatal := platform.AudioStartupError(e); fatal != nil {
				return fatal
			}
			fmt.Fprintln(os.Stderr, "Audio output unavailable:", e)
		} else {
			defer device.Close()
		}
		return platform.ShowFrontend(r, *duration, device)
	}
	var frame *image.RGBA
	path := filepath.Join(*data, "TABLE1.PRG")
	if *pf1 {
		field, err := assets.LoadPartyLand(path)
		if err != nil {
			return err
		}
		frame = field.Framebuffer()
	} else {
		initial, err := assets.LoadInitialPartyLand(path)
		if err != nil {
			return err
		}
		frame = initial.Framebuffer()
	}
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			return err
		}
		err = png.Encode(f, frame)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		return closeErr
	}
	if *pf1 {
		return platform.Show(frame, time.Duration(*duration))
	}
	return platform.ShowInitial(frame, time.Duration(*duration))
}
func main() {
	defer platform.RecoverFatal()

	if err := run(); err != nil && !errors.Is(err, platform.ErrQuit) {
		platform.ReportFatal(err)
		os.Exit(1)
	}
}
