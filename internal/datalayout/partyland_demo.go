package datalayout

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
)

// PartyLandDemoProfile is the official 10-minute DOS demo: Party Land only,
// decoded into the shared TABLE1 address space. TABLE2-4 are not part of it.
const PartyLandDemoProfile = "dos-partyland-10min-demo-v1"

//go:embed partyland_demo.json
var partyLandDemoJSON []byte

type demoRecord struct {
	Source, Size int
	SHA256       string
}

// Generated offline by tools/partyland_demo_layout.py from reviewed evidence.
// It holds offsets, typed jingle values and identities, never original bytes.
var partyLandDemo = func() (l struct {
	Intro struct {
		Regions  []Region
		Pictures []Picture
		Records  map[string]demoRecord
	}
	Table1 struct {
		DecodedSize int `json:"decoded_size"`
		Copies      []recordCopy
		Jingles     []jingleRecord
		Reviewed    []Region
		Records     map[string]demoRecord
	}
}) {
	if err := json.Unmarshal(partyLandDemoJSON, &l); err != nil {
		panic(err)
	}
	return l
}()

func partyLandDemoLinked() linkedLayout {
	t := partyLandDemo.Table1
	return linkedLayout{DecodedSize: t.DecodedSize, Copies: t.Copies, Jingles: t.Jingles}
}

// partyLandDemoProfile maps the canonical read map onto the demo file, or
// describes the decoded result. The two typed jingle differences replace the
// canonical identities; every other hashed consumer record must be equal.
func partyLandDemoProfile(source bool) Profile {
	p := profiles["TABLE1.PRG"]
	p.Profile = PartyLandDemoProfile
	p.Regions = append([]Region(nil), p.Regions...)
	p.Pictures = append([]Picture(nil), p.Pictures...)
	l := partyLandDemoLinked()
	for i := range p.Regions {
		r := &p.Regions[i]
		for _, j := range l.Jingles {
			if r.Offset == j.Destination && r.Size == 3 {
				r.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte{j.Position, j.Repeat, j.Priority}))
			}
		}
		if source {
			r.Offset = l.sourceOffset(r.Offset, r.Size)
		}
	}
	if source {
		for i := range p.Pictures {
			p.Pictures[i].Offset = l.sourceOffset(p.Pictures[i].Offset, 12)
		}
		p.Regions = append(p.Regions, partyLandDemo.Table1.Reviewed...)
		for name, r := range partyLandDemo.Table1.Records {
			p.Regions = append(p.Regions, Region{Offset: r.Source, Size: r.Size, Purpose: "demo record " + name, SHA256: r.SHA256})
		}
	}
	return p
}

func validatePartyLandDemo(name string, data []byte) error {
	switch name {
	case "INTRO.PRG":
		return validateProfile(name, data, partyLandDemoIntro())
	case "TABLE1.PRG":
		return validateProfile(name, data, partyLandDemoProfile(true))
	}
	return fmt.Errorf("%s: not part of the %s installation", name, PartyLandDemoProfile)
}

// The demo INTRO is consumed in place: its pictures and text records have
// their own reviewed positions, described by DecodedFrontendLayout.
func partyLandDemoIntro() Profile {
	p := Profile{Profile: PartyLandDemoProfile, Regions: append([]Region(nil), partyLandDemo.Intro.Regions...), Pictures: partyLandDemo.Intro.Pictures}
	for name, r := range partyLandDemo.Intro.Records {
		p.Regions = append(p.Regions, Region{Offset: r.Source, Size: r.Size, Purpose: "demo record " + name, SHA256: r.SHA256})
	}
	return p
}

func translatePartyLandDemo(name string, data []byte) ([]byte, error) {
	switch name {
	case "INTRO.PRG":
		return append([]byte(nil), data...), nil
	case "TABLE1.PRG":
		return translateRecords(partyLandDemoLinked(), data)
	}
	return nil, fmt.Errorf("%s: not part of the %s installation", name, PartyLandDemoProfile)
}

// demoFrontendLayout reads the demo's own selector text: the welcome and
// availability SHOWTEXT pages (12 rows each) and the sidebar/options record.
func demoFrontendLayout(data []byte) FrontendLayout {
	l := FrontendLayout{Pictures: append([]Picture(nil), partyLandDemo.Intro.Pictures...), StartupLowerY: 139}
	r := partyLandDemo.Intro.Records["SIDEBAR"]
	l.SidebarOffset = r.Source
	for _, name := range []string{"WELCOME_PAGE", "AVAILABLE_PAGE"} {
		r := partyLandDemo.Intro.Records[name]
		var page []string
		for _, row := range bytes.Split(data[r.Source:r.Source+r.Size-1], []byte{0}) {
			page = append(page, string(row))
		}
		l.TextPages = append(l.TextPages, page)
	}
	return l
}

// DetectDemoInstallation accepts only the coherent demo pair. Retail and
// relocated full installations keep using DetectInstallation.
func DetectDemoInstallation(intro, table []byte) (string, error) {
	for name, data := range map[string][]byte{"INTRO.PRG": intro, "TABLE1.PRG": table} {
		if err := validatePartyLandDemo(name, data); err != nil {
			return "", fmt.Errorf("unsupported layout: not the official 10-minute demo: %w", err)
		}
	}
	return PartyLandDemoProfile, nil
}

// PartyLandDemoRecords are the demo-only consumed records outside the shared
// decoded address space, read from a validated demo TABLE1.PRG.
type PartyLandDemoRecords struct {
	PlayersText []byte    // fixed duration label, replaces the mutable player label
	ExpiryTexts [2][]byte // the two expiry scrolls, 255 terminated
}

func DemoRecords(table []byte) (PartyLandDemoRecords, error) {
	var out PartyLandDemoRecords
	if err := validatePartyLandDemo("TABLE1.PRG", table); err != nil {
		return out, err
	}
	record := func(name string) []byte {
		r := partyLandDemo.Table1.Records[name]
		return append([]byte(nil), table[r.Source:r.Source+r.Size]...)
	}
	out.PlayersText = record("PLAYERSTEXT")
	if bytes.IndexByte(out.PlayersText, 0) != len(out.PlayersText)-1 {
		return out, fmt.Errorf("TABLE1.PRG: demo duration label terminator")
	}
	for i, name := range []string{"EXPIRY_TEXT1", "EXPIRY_TEXT2"} {
		text := record(name)
		if bytes.IndexByte(text, 255) != len(text)-1 {
			return out, fmt.Errorf("TABLE1.PRG: demo expiry scroll terminator")
		}
		for _, ch := range text[:len(text)-1] {
			if ch != 1 && (ch < 32 || ch > 96) {
				return out, fmt.Errorf("TABLE1.PRG: demo expiry scroll character outside glyph map")
			}
		}
		out.ExpiryTexts[i] = text
	}
	return out, nil
}
