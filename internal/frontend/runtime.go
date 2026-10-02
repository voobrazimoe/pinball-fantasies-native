package frontend

import (
	"image"
	"os"
	"path/filepath"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"pinballfantasies/internal/tablelogic"
)

type Runtime struct {
	Model       *Model
	View        *View
	Intro, Menu *audio.Module
	Player      *audio.Player
	PCM         []byte
}

func Load(dataDir string, store Store) (*Runtime, error) {
	// Library callers get installation compatibility without user-config writes.
	return LoadConfigured(dataDir, store, nil)
}
func LoadConfigured(dataDir string, store Store, configStore *settings.Store) (*Runtime, error) {
	read := func(name string) ([]byte, error) { return os.ReadFile(filepath.Join(dataDir, name)) }
	b, e := read("INTRO.PRG")
	if e != nil {
		return nil, e
	}
	art, e := assets.DecodeFrontend(b)
	if e != nil {
		return nil, e
	}
	b, e = read("TABLE1.PRG")
	if e != nil {
		return nil, e
	}
	table, e := physics.DecodePartyLand(b)
	if e != nil {
		return nil, e
	}
	tableData := b
	b, e = read("TABLE1.MOD")
	if e != nil {
		return nil, e
	}
	tableMod, e := audio.Decode(b)
	if e != nil {
		return nil, e
	}
	b, e = read("INTRO.MOD")
	if e != nil {
		return nil, e
	}
	intro, e := audio.DecodeFrontend(b)
	if e != nil {
		return nil, e
	}
	b, e = read("MOD2.MOD")
	if e != nil {
		return nil, e
	}
	menu, e := audio.DecodeFrontend(b)
	if e != nil {
		return nil, e
	}
	config := settings.Defaults()
	var err error
	if configStore != nil {
		config, err = configStore.Load()
	} else {
		config, err = (settings.Store{Directory: dataDir}).Load()
	}
	if err != nil {
		return nil, err
	}
	var model *Model
	factory := func(top tablelogic.Decimal) (Session, error) {
		g := partyland.New(table, tableData)
		g.Configure(model.SessionConfig())
		g.SetHighScore(top)
		g.AttachAudio(tableMod)
		return g, nil
	}
	model, e = New(store, factory)
	if e != nil {
		return nil, e
	}
	model.Settings = config
	model.SettingsStore = configStore
	b, e = read("TABLE2.PRG")
	if e != nil {
		return nil, e
	}
	speedData := b
	speedTable, e := physics.DecodeSpeedDevils(b)
	if e != nil {
		return nil, e
	}
	b, e = read("TABLE2.MOD")
	if e != nil {
		return nil, e
	}
	speedMod, e := audio.DecodeSpeedDevils(b)
	if e != nil {
		return nil, e
	}
	model.Factories[1] = func(top tablelogic.Decimal) (Session, error) {
		g := speeddevils.New(speedTable, speedData)
		g.Configure(model.SessionConfig())
		g.SetHighScore(top)
		g.AttachAudio(speedMod)
		return g, nil
	}
	b, e = read("TABLE3.PRG")
	if e != nil {
		return nil, e
	}
	showData := b
	showTable, e := physics.DecodeGameshow(b)
	if e != nil {
		return nil, e
	}
	b, e = read("TABLE3.MOD")
	if e != nil {
		return nil, e
	}
	showMod, e := audio.DecodeGameshow(b)
	if e != nil {
		return nil, e
	}
	model.Factories[2] = func(top tablelogic.Decimal) (Session, error) {
		g := gameshow.New(showTable, showData)
		g.Configure(model.SessionConfig())
		g.SetHighScore(top)
		g.AttachAudio(showMod)
		return g, nil
	}
	b, e = read("TABLE4.PRG")
	if e != nil {
		return nil, e
	}
	stonesData := b
	stonesTable, e := physics.DecodeStones(b)
	if e != nil {
		return nil, e
	}
	b, e = read("TABLE4.MOD")
	if e != nil {
		return nil, e
	}
	stonesMod, e := audio.DecodeStones(b)
	if e != nil {
		return nil, e
	}
	model.Factories[3] = func(top tablelogic.Decimal) (Session, error) {
		g := stones.New(stonesTable, stonesData)
		g.Configure(model.SessionConfig())
		g.SetHighScore(top)
		g.AttachAudio(stonesMod)
		return g, nil
	}
	view := NewView(art, tableData)
	view.SpeedMatrix = presentation.New(2, speedData)
	view.GameshowMatrix = presentation.New(3, showData)
	view.StonesMatrix = presentation.New(4, stonesData)
	view.StonesMatrix.UseSourceMatrixOff()
	return &Runtime{Model: model, View: view, Intro: intro, Menu: menu, Player: audio.New(intro)}, nil
}

// AudioSource identifies a continuous producer, independently of visual modes.
// Selector and selector text share the same player, including its sample phase.
func (r *Runtime) AudioSource() any {
	if r.Model.Session != nil {
		return r.Model.Session
	}
	return r.Player
}

func (r *Runtime) Frame() *image.RGBA { return r.View.Frame(r.Model) }
func (r *Runtime) Update(in Input) error {
	r.PCM = nil
	before := r.Model.Mode
	if e := r.Model.Update(in); e != nil {
		return e
	}
	m := r.Model
	if m.Suspended() || m.Mode == Quit {
		return nil
	}
	if m.Mode == Playing {
		if before == Playing {
			r.PCM = m.Session.PCM()
		}
		return nil
	}
	if m.Session != nil {
		if s, ok := m.Session.(interface{ PresentationAudioSync() }); ok {
			s.PresentationAudioSync()
			r.PCM = m.Session.PCM()
		}
		return nil
	}
	if before >= TableAttract && m.Mode == Selector {
		r.Player = audio.New(r.Menu)
	}
	r.PCM = r.Player.Render(audio.Rate / 60)
	return nil
}
