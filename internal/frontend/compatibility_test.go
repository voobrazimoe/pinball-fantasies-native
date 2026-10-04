package frontend

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"os"
	"path/filepath"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/oracle"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/tablelogic"
	"strings"
	"testing"
)

var runtimeNames = []string{"INTRO.PRG", "INTRO.MOD", "MOD2.MOD", "TABLE1.PRG", "TABLE1.MOD", "TABLE2.PRG", "TABLE2.MOD", "TABLE3.PRG", "TABLE3.MOD", "TABLE4.PRG", "TABLE4.MOD"}

func originalInstallation(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("PF_RUNTIME_DATA")
	if dir == "" {
		dir = os.Getenv("PF_ENGINE_DATA_DIR")
	}
	if dir == "" {
		dir = "../.."
	}
	for _, name := range runtimeNames {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if os.IsNotExist(err) {
			t.Skip("supply pinned originals with PF_RUNTIME_DATA to run compatibility regressions")
		}
		if err != nil {
			t.Fatal(err)
		}
		if err = oracle.Verify(name, b); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
func stageInstallation(t *testing.T, source string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range runtimeNames {
		b, err := os.ReadFile(filepath.Join(source, name))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(dir, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
func loadCompatible(t *testing.T, dir string) *Runtime {
	t.Helper()
	r, err := LoadConfigured(dir, nil, &settings.Store{Directory: t.TempDir(), SeedDirectory: dir})
	if err != nil {
		t.Fatal(err)
	}
	for i, factory := range r.Model.Factories {
		if i == 0 && factory == nil {
			factory = r.Model.Factory
		}
		if factory == nil {
			t.Fatalf("table %d missing factory", i+1)
		}
		if _, err := factory(tablelogic.Decimal{}); err != nil {
			t.Fatalf("table %d: %v", i+1, err)
		}
	}
	return r
}
func TestRuntimeOriginalsAndOptionalSettings(t *testing.T) {
	source := originalInstallation(t)
	for _, tc := range []struct {
		name string
		cfg  []byte
	}{
		{"absent", nil}, {"modified legacy", []byte{1, 1, 2, 1, 1, 1}},
		{"malformed legacy", []byte{255, 3}},
		{"unsupported PFNC", []byte{'P', 'F', 'N', 'C', 99, 5, 0, 0, 1, 0, 0}}, {"native PFNC", settings.Defaults().Bytes()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := stageInstallation(t, source)
			if tc.cfg != nil {
				if err := os.WriteFile(filepath.Join(dir, "PINBALL.CFG"), tc.cfg, 0600); err != nil {
					t.Fatal(err)
				}
			}
			r := loadCompatible(t, dir)
			expected := settings.Defaults()
			if tc.name == "modified legacy" {
				var err error
				expected, err = settings.Decode(tc.cfg)
				if err != nil {
					t.Fatal(err)
				}
			}
			if r.Model.Settings != expected {
				t.Fatalf("settings: got %+v want %+v", r.Model.Settings, expected)
			}
		})
	}
}
func TestRuntimeNativeStatePrecedesLegacySeed(t *testing.T) {
	source := originalInstallation(t)
	dir := stageInstallation(t, source)
	if err := os.WriteFile(filepath.Join(dir, "PINBALL.CFG"), []byte{1, 1, 2, 1, 1, 1}, 0600); err != nil {
		t.Fatal(err)
	}
	store := &settings.Store{Directory: t.TempDir(), SeedDirectory: dir}
	want := settings.Defaults()
	want.ScrollMode = settings.ScrollOff
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	r, err := LoadConfigured(dir, nil, store)
	if err != nil {
		t.Fatal(err)
	}
	if r.Model.Settings != want {
		t.Fatal("PFNC native state did not override legacy seed")
	}
}

func TestRuntimeUnusedPRGHeaderAndStrictOracle(t *testing.T) {
	source := originalInstallation(t)
	dir := stageInstallation(t, source)
	for _, name := range runtimeNames {
		if !strings.HasSuffix(name, ".PRG") {
			continue
		}
		path := filepath.Join(dir, name)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := sha256.Sum256(b)
		// Source audit: no runtime reader consumes MZ/container bytes. All reads
		// are the explicit profile regions/IFFs, with the earliest at INTRO 0x6b70
		// or table glyph stores >0x6000. Byte zero is outside every consumed range.
		b[0] ^= 1
		if before == sha256.Sum256(b) {
			t.Fatal("mutation did not alter identity")
		}
		if oracle.Verify(name, b) == nil {
			t.Fatal("oracle accepted changed executable header")
		}
		if err = os.WriteFile(path, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	loadCompatible(t, dir)
}
func TestRuntimeRejectsConsumedDamage(t *testing.T) {
	source := originalInstallation(t)
	cases := []struct {
		name   string
		at     int
		change func([]byte) []byte
	}{
		{"unsupported moved layout", 0, func(b []byte) []byte { return append([]byte{0}, b...) }},
		{"FORM anchor", 336944, func(b []byte) []byte { b[336944] ^= 1; return b }},
		{"PBM geometry", 336944, func(b []byte) []byte { binary.BigEndian.PutUint16(b[336944+20:], 319); return b }},
		{"truncated physics", 0, func(b []byte) []byte { return b[:0x77530] }},
		{"flipper descriptor", 0x20690, func(b []byte) []byte { b[0x20690+6] ^= 1; return b }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := stageInstallation(t, source)
			path := filepath.Join(dir, "TABLE1.PRG")
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, tc.change(b), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err = LoadConfigured(dir, nil, &settings.Store{Directory: t.TempDir()}); err == nil || !strings.Contains(err.Error(), "unsupported layout") {
				t.Fatalf("expected unsupported layout, got %v", err)
			}
		})
	}
}
func TestRuntimeAudioPayloadIsNotOracle(t *testing.T) {
	source := originalInstallation(t)
	dir := stageInstallation(t, source)
	decoders := map[string]func([]byte) (*audio.Module, error){"INTRO.MOD": audio.DecodeIntro, "MOD2.MOD": audio.DecodeMenu, "TABLE1.MOD": audio.Decode, "TABLE2.MOD": audio.DecodeSpeedDevils, "TABLE3.MOD": audio.DecodeGameshow, "TABLE4.MOD": audio.DecodeStones}
	for name, decode := range decoders {
		path := filepath.Join(dir, name)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		m, err := decode(b)
		if err != nil {
			t.Fatal(err)
		}
		// First sample PCM byte, after the validated pattern extent, has no layout role.
		off := 1084 + len(m.Patterns)*1024
		b[off] ^= 1
		if oracle.Verify(name, b) == nil {
			t.Fatal("changed PCM passed strict oracle")
		}
		if err = os.WriteFile(path, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	loadCompatible(t, dir)
}
func TestRuntimeOpaqueArtworkCanChange(t *testing.T) {
	source := originalInstallation(t)
	b, err := os.ReadFile(filepath.Join(source, "TABLE1.PRG"))
	if err != nil {
		t.Fatal(err)
	}
	// Foreground is a raw 40-byte by 576-row bitplane, with no pointers/control.
	before, err := assets.DecodeInitialPartyLand(b)
	if err != nil {
		t.Fatal(err)
	}
	b[0x300b0] ^= 1
	after, err := assets.DecodeInitialPartyLand(b)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before.Playfield.Indices, after.Playfield.Indices) {
		t.Fatal("foreground changed playfield")
	}
}

func TestHostsShareRuntimeBoundary(t *testing.T) {
	// Linux/Windows use the CLI loader; C ABI hosts use engine.Load.
	// Both routes must reach the same shared runtime compatibility decision.
	for _, tc := range []struct{ path, call string }{
		{"../../cmd/pinballfantasies/main.go", "frontend.LoadConfigured("},
		{"../../internal/engine/engine.go", "frontend.LoadConfigured("},
		{"../../internal/frontend/runtime.go", "datalayout.PreparePRG("},
		{"../../cmd/pfengine/main.go", "engine.Load("},
	} {
		b, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), tc.call) {
			t.Fatalf("shared compatibility call missing in %s: %s", tc.path, tc.call)
		}
	}
}
