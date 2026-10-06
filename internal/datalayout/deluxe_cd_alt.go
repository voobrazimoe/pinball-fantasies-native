package datalayout

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed deluxe_cd_alt.json
var deluxeCDAltJSON []byte

type factorySeed struct {
	Source, Destination int
	Initials            [4]string
	SHA256              string
}

// Only INTRO addresses and factory identity differ. TABLE copy spans, cue
// records, geometry and reviewed Stones role mapping have a single definition.
var deluxeAltMetadata = func() struct {
	Intro deluxeLayout
	Seeds map[string]factorySeed
} {
	var out struct {
		Intro deluxeLayout
		Seeds map[string]factorySeed
	}
	if err := json.Unmarshal(deluxeCDAltJSON, &out); err != nil {
		panic(err)
	}
	return out
}()

func deluxeAltLayout(name string) (deluxeLayout, bool) {
	if name == "INTRO.PRG" {
		return deluxeAltMetadata.Intro, true
	}
	l, ok := deluxeLayouts[name]
	if !ok {
		return l, false
	}
	l.SourceProfile.Regions = append([]Region(nil), l.SourceProfile.Regions...)
	l.SourceProfile.Profile = DeluxeCDAltProfile
	seed := deluxeAltMetadata.Seeds[name]
	for i := range l.SourceProfile.Regions {
		if l.SourceProfile.Regions[i].Purpose == "matrix record HI_SCORE_LIST" {
			l.SourceProfile.Regions[i].SHA256 = seed.SHA256
		}
	}
	return l, true
}

func validateDeluxeAlt(name string, data []byte) error {
	l, ok := deluxeAltLayout(name)
	if !ok {
		return fmt.Errorf("%s: no supported data layout", name)
	}
	if err := validateProfile(name, data, l.SourceProfile); err != nil {
		return err
	}
	if seed, ok := deluxeAltMetadata.Seeds[name]; ok {
		return validateFactoryInitials(data, seed.Source, seed.Initials)
	}
	return nil
}

func validateFactoryInitials(data []byte, at int, initials [4]string) error {
	if at < 0 || at > len(data)-64 {
		return fmt.Errorf("truncated factory scores")
	}
	for i, want := range initials {
		if len(want) != 3 || string(data[at+i*16+12:at+i*16+15]) != want {
			return fmt.Errorf("incompatible factory initials entry %d", i)
		}
	}
	return nil
}

func translateDeluxeAlt(name string, data []byte) ([]byte, error) {
	l, ok := deluxeAltLayout(name)
	if !ok {
		return nil, fmt.Errorf("%s: no supported data layout", name)
	}
	return translateDeluxeLayout(l, data)
}

func deluxeAltDecodedProfile(name string) Profile {
	p := deluxeDecodedProfile(name)
	p.Profile = DeluxeCDAltProfile
	p.Regions = append([]Region(nil), p.Regions...)
	if seed, ok := deluxeAltMetadata.Seeds[name]; ok {
		// The decoder contract retains the actual C seed, including every initial.
		for i := range p.Regions {
			if p.Regions[i].Purpose == "matrix record HI_SCORE_LIST" {
				p.Regions[i].SHA256 = seed.SHA256
			}
		}
	}
	return p
}

// FactoryInitials reads the validated typed seed in the common decoder address
// space. Numeric native defaults retain the existing INTRO factory contract;
// TABLE3's attract list has different score amounts even in canonical A.
func FactoryInitials(name string, data []byte) ([4][3]byte, error) {
	var out [4][3]byte
	if err := ValidateDecoded(name, data); err != nil {
		return out, err
	}
	for _, r := range profiles[name].Regions {
		if r.Purpose != "matrix record HI_SCORE_LIST" {
			continue
		}
		for i := range out {
			copy(out[i][:], data[r.Offset+i*16+12:r.Offset+i*16+15])
		}
		return out, nil
	}
	return out, fmt.Errorf("%s: no factory score record", name)
}
