//go:build demodev

package partyland

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/oracle"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
)

// Experimental admission pins the same five-file input as the existing
// candidate. It does not register a supported public installation profile.
var experimentalDemoFiles = map[string]struct {
	size     int
	identity string
}{
	"INTRO.PRG":  {347054, "05bdba35e0a9a31a87b944428ddad27a00ba8983e97963290ab2ee1f57910fa3"},
	"INTRO.MOD":  {252870, "f36beae00efec1dd9e1c4e977bea264b7ec41ab18f258528ab66577a9ec66613"},
	"MOD2.MOD":   {55394, "aa5003c275b494062f37f44e8c77105b8a420555f4bd6ff53d7698f89c540f21"},
	"TABLE1.PRG": {537190, "44b8f4b76ee16c47cda26904681e83e8f4420974bcea249d099790bfbd7369e3"},
	"TABLE1.MOD": {210760, "a0877e4372abe64b70d9e361bf257ea5a84c948771f0eace3433d5f6399060b5"},
}

func LoadExperimentalDemo(demoDir, canonicalDir string) (*ExperimentalDemo, error) {
	inputs := make(map[string][]byte)
	for name, want := range experimentalDemoFiles {
		b, err := os.ReadFile(filepath.Join(demoDir, name))
		if err != nil {
			return nil, err
		}
		if len(b) != want.size || fmt.Sprintf("%x", sha256.Sum256(b)) != want.identity {
			return nil, fmt.Errorf("%s: not the verified official 10-minute demo input", name)
		}
		inputs[name] = b
	}
	a, err := os.ReadFile(filepath.Join(canonicalDir, "TABLE1.PRG"))
	if err != nil {
		return nil, err
	}
	if err = oracle.Verify("TABLE1.PRG", a); err != nil {
		return nil, fmt.Errorf("explicit canonical-A asset input: %w", err)
	}
	a, err = datalayout.PreparePRGForProfile(datalayout.RetailProfile, "TABLE1.PRG", a)
	if err != nil {
		return nil, err
	}
	table, err := physics.DecodePartyLand(a)
	if err != nil {
		return nil, err
	}
	module, err := audio.Decode(inputs["TABLE1.MOD"])
	if err != nil {
		return nil, err
	}
	g := New(table, a)
	g.Configure(settings.Legacy())
	// Own mutable presentation data before any demo overlay is applied.
	texts := make(map[string][]byte)
	for k, v := range g.Display.Content.Texts {
		texts[k] = append([]byte(nil), v...)
	}
	g.Display.Content.Texts = texts
	fonts := make(map[string]int)
	for k, v := range g.Display.Content.Fonts {
		fonts[k] = v
	}
	g.Display.Content.Fonts = fonts
	d := connected(g)
	b := inputs["TABLE1.PRG"]
	for _, admit := range []func([]byte, []byte) error{d.admitGameplay, d.admitHighScore, d.admitScoredDrain, d.admitInfo} {
		if err = admit(b, a); err != nil {
			return nil, err
		}
	}
	d.toucherOperands = true
	// No canonical callback may be consumed by the staged runtime.
	g.Physics.OnEvent = nil
	g.Physics.BeforeTargets = nil
	g.Physics.AfterTargets = nil
	g.Physics.BeforeLate = nil
	p := audio.New(module)
	p.Jump = func(next int) int {
		if g.Audio.JumpCount == 1 {
			return int(g.Audio.ReturnPosition)
		}
		return next
	}
	return &ExperimentalDemo{core: d, player: p}, nil
}
