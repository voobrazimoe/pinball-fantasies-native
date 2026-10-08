package datalayout

import (
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
)

//go:embed deluxe_cd.json
var deluxeCDJSON []byte

type selectorRecord struct {
	Destination, Source int
	Raw                 uint16
	NativeID            uint16 `json:"native_id"`
	Role                string
}
type deluxeLayout struct {
	linkedLayout
	SourceProfile Profile `json:"source_profile"`
	Selectors     []selectorRecord
}

var deluxeLayouts = func() map[string]deluxeLayout {
	var out map[string]deluxeLayout
	if err := json.Unmarshal(deluxeCDJSON, &out); err != nil {
		panic(err)
	}
	return out
}()

func validateDeluxe(name string, data []byte) error {
	l, ok := deluxeLayouts[name]
	if !ok {
		return fmt.Errorf("%s: no supported data layout", name)
	}
	return validateProfile(name, data, l.SourceProfile)
}

func translateDeluxe(name string, data []byte) ([]byte, error) {
	l, ok := deluxeLayouts[name]
	if !ok {
		return nil, fmt.Errorf("%s: no supported data layout", name)
	}
	return translateDeluxeLayout(l, data)
}

func translateDeluxeLayout(l deluxeLayout, data []byte) ([]byte, error) {
	decoded, err := translateRecords(l.linkedLayout, data)
	if err != nil {
		return nil, err
	}
	// Typed identity conversion. Raw DOS selectors never reach native consumers.
	// Rectangle copies deliberately exclude every selector word.
	for _, s := range l.Selectors {
		if s.Source < 0 || s.Source > len(data)-2 || s.Destination < 0 || s.Destination > len(decoded)-2 {
			return nil, fmt.Errorf("bounded area selector %s", s.Role)
		}
		if binary.LittleEndian.Uint16(data[s.Source:]) != s.Raw {
			return nil, fmt.Errorf("unknown area selector %s", s.Role)
		}
		binary.LittleEndian.PutUint16(decoded[s.Destination:], s.NativeID)
	}
	return decoded, nil
}

func deluxeDecodedProfile(name string) Profile {
	p := profiles[name]
	p.Profile = DeluxeCDProfile
	p.Pictures = append([]Picture(nil), p.Pictures...)
	if name == "INTRO.PRG" {
		p.Pictures[7].Height = 123
		p.Pictures[8].Height = 117
	}
	if name == "TABLE1.PRG" {
		p.Regions = linkedProfile(name, false).Regions // shared priority-zero cue contract
	}
	return p
}

// FrontendLayout describes presentation roles in the shared decoded address
// space. Geometry and placement are data; the renderer has no edition identity.
type FrontendLayout struct {
	Pictures      []Picture
	StartupLowerY int
	SidebarOffset int        // selector sidebar, then options sidebar, 120 bytes each
	TextPages     [][]string // non-nil: this edition's own SHOWTEXT cycle
}

func DecodedFrontendLayout(data []byte) (FrontendLayout, error) {
	if validateProfile("INTRO.PRG", data, partyLandDemoIntro()) == nil {
		return demoFrontendLayout(data), nil
	}
	for _, p := range []Profile{profiles["INTRO.PRG"], deluxeDecodedProfile("INTRO.PRG")} {
		if err := validateProfile("INTRO.PRG", data, p); err == nil {
			lowerY := 139
			if p.Profile == DeluxeCDProfile {
				lowerY = 123
			}
			return FrontendLayout{Pictures: append([]Picture(nil), p.Pictures...), StartupLowerY: lowerY, SidebarOffset: 233806}, nil
		}
	}
	return FrontendLayout{}, fmt.Errorf("INTRO.PRG: unsupported decoded presentation layout")
}
