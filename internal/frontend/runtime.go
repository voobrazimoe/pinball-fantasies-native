package frontend

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"pinballfantasies/internal/tablelogic"
	"strings"
)

type Runtime struct {
	Model       *Model
	View        *View
	Intro, Menu *audio.Module
	Player      *audio.Player
	PCM         []byte
	ProfileID   string
}

var runtimeNamesRequired = []string{"INTRO.PRG", "INTRO.MOD", "MOD2.MOD", "TABLE1.PRG", "TABLE1.MOD", "TABLE2.PRG", "TABLE2.MOD", "TABLE3.PRG", "TABLE3.MOD", "TABLE4.PRG", "TABLE4.MOD"}

func Load(dataDir string, store Store) (*Runtime, error) {
	// Library callers get installation compatibility without user-config writes.
	return LoadConfigured(dataDir, store, nil)
}
func LoadConfigured(dataDir string, store Store, configStore *settings.Store) (*Runtime, error) {
	// Read required inputs before decoding so incomplete installs keep their
	// normal filename error. Adoption and relaunch re-detect the same profile.
	inputs := make(map[string][]byte)
	prgs := make(map[string][]byte)
	for _, name := range runtimeNamesRequired {
		data, err := os.ReadFile(filepath.Join(dataDir, name))
		if err != nil {
			return nil, err
		}
		inputs[name] = data
		if strings.HasSuffix(name, ".PRG") {
			prgs[name] = data
		}
	}
	profileID, err := datalayout.DetectInstallation(prgs)
	if err != nil {
		return nil, err
	}
	prepared := make(map[string][]byte)
	read := func(name string) ([]byte, error) {
		data := inputs[name]
		if strings.HasSuffix(name, ".PRG") {
			if cached, ok := prepared[name]; ok {
				return cached, nil
			}
			decoded, err := datalayout.PreparePRGForProfile(profileID, name, data)
			if err == nil {
				prepared[name] = decoded
			}
			return decoded, err
		}
		return data, nil
	}
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
		return nil, fmt.Errorf("TABLE1.MOD: %w", e)
	}
	b, e = read("INTRO.MOD")
	if e != nil {
		return nil, e
	}
	intro, e := audio.DecodeIntro(b)
	if e != nil {
		return nil, fmt.Errorf("INTRO.MOD: %w", e)
	}
	b, e = read("MOD2.MOD")
	if e != nil {
		return nil, e
	}
	menu, e := audio.DecodeMenu(b)
	if e != nil {
		return nil, fmt.Errorf("MOD2.MOD: %w", e)
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
	var model *Model
	factory := func(top tablelogic.Decimal) (Session, error) {
		g := partyland.New(table, tableData)
		g.Configure(model.SessionConfig())
		g.SetHighScore(top)
		g.AttachAudio(tableMod)
		return g, nil
	}
	var factoryScores [4]Scores
	for i := range factoryScores {
		name := fmt.Sprintf("TABLE%d.PRG", i+1)
		decoded, err := read(name)
		if err != nil {
			return nil, err
		}
		initials, err := datalayout.FactoryInitials(name, decoded)
		if err != nil {
			return nil, err
		}
		factoryScores[i] = Defaults(i + 1)
		for rank := range factoryScores[i] {
			factoryScores[i][rank].Name = initials[rank]
		}
	}
	switch s := store.(type) {
	case FileStore:
		store = installationScoreStore{FileStore: s, defaults: factoryScores}
	case *FileStore:
		store = installationScoreStore{FileStore: *s, defaults: factoryScores}
	}
	model, e = newWithDefaults(store, factory, factoryScores)
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
		return nil, fmt.Errorf("TABLE2.MOD: %w", e)
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
		return nil, fmt.Errorf("TABLE3.MOD: %w", e)
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
		return nil, fmt.Errorf("TABLE4.MOD: %w", e)
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
	return &Runtime{ProfileID: profileID, Model: model, View: view, Intro: intro, Menu: menu, Player: audio.New(intro)}, nil
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
	if m.sessionSynced {
		r.PCM = m.Session.PCM()
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

// FramePresentation scopes a render-only override to this retrieval. It cannot
// affect Update, source camera arithmetic, settings writes or future sessions.
func (r *Runtime) FramePresentation(full bool) *image.RGBA {
	var p *physics.Game
	switch g := r.Model.Session.(type) {
	case *partyland.Game:
		p = g.Physics
	case *speeddevils.Game:
		p = g.Physics
	case *gameshow.Game:
		p = g.Physics
	case *stones.Game:
		p = g.Physics
	}
	if p != nil {
		previous := p.PresentationFullTable
		p.PresentationFullTable = full
		defer func() { p.PresentationFullTable = previous }()
	}
	return r.View.framePresentation(r.Model, full)
}
