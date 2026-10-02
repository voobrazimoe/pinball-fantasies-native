// Package settings separates native configuration from the original DOS record.
package settings

import (
	"fmt"
	"os"
	"path/filepath"
)

// ScrollMode extends the three DOS camera modes with native presentation OFF.
type ScrollMode uint8

const (
	ScrollHard ScrollMode = iota
	ScrollMedium
	ScrollSoft
	ScrollOff
)

// Config contains native settings only. DOS S_MODE belongs to LegacyRecord.
type Config struct {
	Balls, Angle      byte
	ScrollMode        ScrollMode
	Music, Resolution byte
}

type LegacyRecord struct{ Balls, Angle, Scrolling, Music, Resolution, Mode byte }

func DecodeLegacy(b []byte) (LegacyRecord, error) {
	if len(b) != 6 {
		return LegacyRecord{}, fmt.Errorf("DOS PINBALL.CFG requires six bytes")
	}
	for i, v := range b {
		max := byte(1)
		if i == 2 {
			max = 2
		}
		if v > max {
			return LegacyRecord{}, fmt.Errorf("invalid DOS field %d: %d", i, v)
		}
	}
	return LegacyRecord{b[0], b[1], b[2], b[3], b[4], b[5]}, nil
}
func (r LegacyRecord) Native() Config {
	return Config{r.Balls, r.Angle, ScrollMode(r.Scrolling), r.Music, r.Resolution}
}
func Defaults() Config { return Config{ScrollMode: ScrollMedium} }

// Legacy preserves the accepted high-mode, low-angle, soft-scroll oracles.
func Legacy() Config { return Config{Angle: 1, ScrollMode: ScrollSoft, Resolution: 1} }

// Native v1: PFNC, version=1, payload length=5, then the five native fields.
// Future versions may define a different payload; unknown versions are rejected.
func (c Config) Bytes() []byte {
	return []byte{'P', 'F', 'N', 'C', 1, 5, c.Balls, c.Angle, byte(c.ScrollMode), c.Music, c.Resolution}
}
func (c Config) Validate() error {
	if c.Balls > 1 || c.Angle > 1 || c.ScrollMode > ScrollOff || c.Music > 1 || c.Resolution > 1 {
		return fmt.Errorf("invalid native settings")
	}
	return nil
}
func Decode(b []byte) (Config, error) {
	if len(b) >= 4 && string(b[:4]) == "PFNC" {
		if len(b) < 6 {
			return Defaults(), fmt.Errorf("truncated native header")
		}
		if b[4] != 1 {
			return Defaults(), fmt.Errorf("unsupported native config version %d", b[4])
		}
		if b[5] != 5 || len(b) != 11 {
			return Defaults(), fmt.Errorf("invalid native payload length")
		}
		c := Config{b[6], b[7], ScrollMode(b[8]), b[9], b[10]}
		if e := c.Validate(); e != nil {
			return Defaults(), e
		}
		return c, nil
	}
	r, e := DecodeLegacy(b)
	if e != nil {
		return Defaults(), e
	}
	return r.Native(), nil // Both DOS COLOR and MONO import to native COLOR.
}
func (c *Config) Cycle(row int) {
	if row == 2 {
		c.ScrollMode = (c.ScrollMode + 1) % 4
		return
	}
	fields := [5]*byte{&c.Balls, &c.Angle, nil, &c.Music, &c.Resolution}
	if row < 0 || row >= len(fields) {
		return
	}
	if field := fields[row]; field != nil {
		*field = (*field + 1) % 2
	}
}
func (c Config) Values() [5]string {
	return [5]string{[]string{"3", "5"}[c.Balls], []string{"HIGH  ", "LOW   "}[c.Angle], []string{"HARD  ", "MEDIUM", "SOFT  ", "OFF   "}[c.ScrollMode], []string{"ON ", "OFF"}[c.Music], []string{"NORMAL", "HIGH  "}[c.Resolution]}
}

// FieldHeight is the source camera viewport, even when native OFF hides it.
func (c Config) FieldHeight() int {
	if c.Resolution == 0 {
		return 207
	}
	return 317
}
func (c Config) RenderHeight() int {
	if c.ScrollMode == ScrollOff {
		return 576
	}
	return c.FieldHeight()
}
func (c Config) MatrixY() int {
	if c.ScrollMode == ScrollOff {
		return 0
	}
	return c.FieldHeight()
}
func (c Config) TableY() int {
	if c.ScrollMode == ScrollOff {
		return 33
	}
	return 0
}

// OFF retains SOFT's internal camera arithmetic; it only changes composition.
func (c Config) ScrollFactor() int16 { return [...]int16{20, 11, 9, 9}[c.ScrollMode] }

// MonoDAC implements pelle_2_bw, on six-bit DAC units.
func MonoDAC(p [768]byte) [768]byte {
	for i := 0; i < len(p); i += 3 {
		v := byte((int(p[i]>>2) + int(p[i+1]>>2) + int(p[i+2]>>2)) / 3)
		v = (v << 2) | (v >> 4)
		p[i], p[i+1], p[i+2] = v, v, v
	}
	return p
}

// MonoRGB implements COLOR_2_BW on already recalculated light RGB triplets.
func MonoRGB(b []byte) []byte {
	b = append([]byte(nil), b...)
	for i := 0; i+2 < len(b); i += 3 {
		v := byte((int(b[i]) + int(b[i+1]) + int(b[i+2])) / 3)
		b[i], b[i+1], b[i+2] = v, v, v
	}
	return b
}

type Store struct{ Directory, SeedDirectory string }

func (s Store) Load() (Config, error) {
	b, e := os.ReadFile(filepath.Join(s.Directory, "PINBALL.CFG"))
	if os.IsNotExist(e) && s.SeedDirectory != "" {
		b, e = os.ReadFile(filepath.Join(s.SeedDirectory, "PINBALL.CFG"))
	}
	if os.IsNotExist(e) {
		return Defaults(), nil
	}
	if e != nil {
		return Config{}, e
	}
	c, e := Decode(b)
	if e != nil {
		return Defaults(), nil
	}
	return c, nil
}
func (s Store) Save(c Config) error {
	if _, e := Decode(c.Bytes()); e != nil {
		return e
	}
	if e := os.MkdirAll(s.Directory, 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(s.Directory, ".options-*")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	_, e = f.Write(c.Bytes())
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e != nil {
		return e
	}
	if ce != nil {
		return ce
	}
	return os.Rename(f.Name(), filepath.Join(s.Directory, "PINBALL.CFG"))
}
