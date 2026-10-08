//go:build linux || windows

package platform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"pinballfantasies/internal/demodata"
	"pinballfantasies/internal/frontend"
)

// ErrQuit reports that the player closed the launch choice.
var ErrQuit = errors.New("quit")

type launchChoice int

const (
	launchImport launchChoice = iota
	launchDemo
	launchQuit
)

// ChooseDataDir finds the game data for the interactive frontend. An explicit
// directory, a personal bundle, a game beside the executable or a previously
// imported game start directly. Otherwise the player either imports the full
// game into native storage or plays the bundled 10-minute demo.
func ChooseDataDir(value string) (string, func(), error) {
	if value != "" || len(personalPayload) != 0 {
		return PrepareDataDir(value)
	}
	if root, err := PortableRoot(); err == nil {
		if _, err := os.Stat(filepath.Join(root, "INTRO.PRG")); err == nil {
			return root, func() {}, nil
		}
	}
	storage, err := UserStorageDir()
	if err != nil {
		return "", nil, err
	}
	imported := filepath.Join(storage, "Data")
	// The import was validated as a whole; startup only checks it is still there.
	if complete(imported) {
		return imported, func() {}, nil
	}
	return chooseLaunch(imported, askLaunch, pickFolder, showMessage)
}

func chooseLaunch(imported string, ask func() launchChoice, pick func() (string, bool), show func(string)) (string, func(), error) {
	for {
		switch ask() {
		case launchDemo:
			return demodata.Extract()
		case launchImport:
			source, ok := pick()
			if !ok {
				continue
			}
			if err := importGame(source, imported); err != nil {
				show("These files could not be imported.\n\n" + err.Error())
				continue
			}
			return imported, func() {}, nil
		default:
			return "", nil, ErrQuit
		}
	}
}

func complete(dir string) bool {
	for _, name := range frontend.FullNamesRequired() {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			return false
		}
	}
	return true
}

// fullGame accepts only an installation the frontend loads as the full game.
func fullGame(dir string) error {
	r, err := frontend.Load(dir, frontend.FileStore{Directory: dir})
	if err != nil {
		return err
	}
	if r.ProfileID == frontend.DemoProfileID {
		return fmt.Errorf("this folder holds the 10-minute demo; choose the full game folder with INTRO.PRG and TABLE1-4")
	}
	return nil
}

// importGame validates the source, copies its files into a staging folder,
// validates the copies and only then replaces the previous import.
func importGame(source, destination string) error {
	if err := fullGame(source); err != nil {
		return err
	}
	parent := filepath.Dir(destination)
	stage, err := os.MkdirTemp(parent, ".import-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	// Settings and high-score files are optional seeds, as beside the executable.
	required := frontend.FullNamesRequired()
	names := append(required, "PINBALL.CFG", "TABLE1.HI", "TABLE2.HI", "TABLE3.HI", "TABLE4.HI")
	for i, name := range names {
		b, err := os.ReadFile(filepath.Join(source, name))
		if errors.Is(err, os.ErrNotExist) && i >= len(required) {
			continue
		}
		if err == nil {
			err = os.WriteFile(filepath.Join(stage, name), b, 0600)
		}
		if err != nil {
			return err
		}
	}
	if err := fullGame(stage); err != nil {
		return err
	}
	previous := destination + ".previous"
	os.RemoveAll(previous)
	if err := os.Rename(destination, previous); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(stage, destination); err != nil {
		os.Rename(previous, destination)
		return err
	}
	os.RemoveAll(previous)
	return nil
}

// SDL message boxes do not wrap, so the text carries its own line breaks.
const launchText = "Pinball Fantasies needs the original DOS game files.\n\n" +
	"Import the full game from the folder that holds INTRO.PRG and TABLE1-4,\n" +
	"or play the official 10-minute Party Land demo that comes with this build."
