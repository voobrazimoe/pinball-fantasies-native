package frontend

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/tablelogic"
)

// DemoNamesRequired is the official 10-minute demo's runtime fingerprint.
// Launcher, drivers, TIMER.BIN and settings files are not runtime roles.
var DemoNamesRequired = []string{"INTRO.PRG", "INTRO.MOD", "MOD2.MOD", "TABLE1.PRG", "TABLE1.MOD"}

// DemoProfileID marks a runtime loaded from the official 10-minute demo.
const DemoProfileID = datalayout.PartyLandDemoProfile

// loadDemo runs the demo's own INTRO: startup, selector with its NOT
// AVAILABLE advertising cards, text pages and options. Party Land is the only
// table; TABLE2-4 have no factory. High scores are volatile factory values.
func loadDemo(dataDir string, configStore *settings.Store) (*Runtime, error) {
	inputs := make(map[string][]byte)
	for _, name := range DemoNamesRequired {
		data, err := os.ReadFile(filepath.Join(dataDir, name))
		if err != nil {
			return nil, err
		}
		inputs[name] = data
	}
	id, err := datalayout.DetectDemoInstallation(inputs["INTRO.PRG"], inputs["TABLE1.PRG"])
	if err != nil {
		return nil, err
	}
	introData, err := datalayout.PreparePRGForProfile(id, "INTRO.PRG", inputs["INTRO.PRG"])
	if err != nil {
		return nil, err
	}
	art, err := assets.DecodeFrontend(introData)
	if err != nil {
		return nil, err
	}
	data, err := datalayout.PreparePRGForProfile(id, "TABLE1.PRG", inputs["TABLE1.PRG"])
	if err != nil {
		return nil, err
	}
	records, err := datalayout.DemoRecords(inputs["TABLE1.PRG"])
	if err != nil {
		return nil, err
	}
	table, err := physics.DecodePartyLand(data)
	if err != nil {
		return nil, err
	}
	module, err := audio.Decode(inputs["TABLE1.MOD"])
	if err != nil {
		return nil, fmt.Errorf("TABLE1.MOD: %w", err)
	}
	intro, err := audio.DecodeIntro(inputs["INTRO.MOD"])
	if err != nil {
		return nil, fmt.Errorf("INTRO.MOD: %w", err)
	}
	menu, err := audio.DecodeMenu(inputs["MOD2.MOD"])
	if err != nil {
		return nil, fmt.Errorf("MOD2.MOD: %w", err)
	}
	config := settings.Defaults()
	if configStore != nil {
		config, err = configStore.Load()
	} else {
		config, err = (settings.Store{Directory: dataDir}).Load()
	}
	if err != nil {
		return nil, err
	}
	var defaults [4]Scores
	for i := range defaults {
		defaults[i] = Defaults(i + 1)
	}
	initials, err := datalayout.FactoryInitials("TABLE1.PRG", data)
	if err != nil {
		return nil, err
	}
	for rank := range defaults[0] {
		defaults[0][rank].Name = initials[rank]
	}
	var model *Model
	factory := func(top tablelogic.Decimal) (Session, error) {
		// Every table entry is a fresh table lifetime: timer 0, not expired.
		g := partyland.NewDemo(table, data, partyland.DemoInputs{PlayersText: records.PlayersText, ExpiryTexts: records.ExpiryTexts})
		g.Configure(model.SessionConfig())
		g.SetHighScore(top)
		g.AttachAudio(module)
		return g, nil
	}
	// No HI load/save is proved for the demo: scores live for this run only.
	model, err = newWithDefaults(nil, factory, defaults)
	if err != nil {
		return nil, err
	}
	model.Demo = true
	model.Settings = config
	model.SettingsStore = configStore
	return &Runtime{ProfileID: id, Model: model, View: NewView(art, data), Intro: intro, Menu: menu, Player: audio.New(intro)}, nil
}

// A folder without any of TABLE2-4.PRG is offered to the demo detector; a
// partial full installation keeps its ordinary missing-file error.
func missingFullInstallation(err error) bool { return errors.Is(err, fs.ErrNotExist) }
func demoFolder(dataDir string) bool {
	for _, name := range []string{"TABLE2.PRG", "TABLE3.PRG", "TABLE4.PRG"} {
		if _, err := os.Stat(filepath.Join(dataDir, name)); !errors.Is(err, fs.ErrNotExist) {
			return false
		}
	}
	return true
}
