package datalayout

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"pinballfantasies/internal/tablelogic"
)

//go:embed powerpack.json
var powerpackJSON []byte

type recordCopy struct{ Destination, Source, Size int }
type jingleRecord struct {
	Role                       string
	Destination, Source        int
	Position, Repeat, Priority uint8
}
type linkedLayout struct {
	DecodedSize int `json:"decoded_size"`
	Copies      []recordCopy
	Jingles     []jingleRecord
}

// This descriptor contains reviewed addresses and typed semantics only. It is
// generated offline from evidence; production never searches executable bytes.
var powerpackLayouts = func() map[string]linkedLayout {
	var layouts map[string]linkedLayout
	if err := json.Unmarshal(powerpackJSON, &layouts); err != nil {
		panic(err)
	}
	return layouts
}()

func (l linkedLayout) sourceOffset(at, size int) int {
	for _, c := range l.Copies {
		if at >= c.Destination && size <= c.Size && at-c.Destination <= c.Size-size {
			return c.Source + at - c.Destination
		}
	}
	panic(fmt.Sprintf("unmapped consumed record %#x/%d", at, size))
}

func linkedProfile(name string, source bool) Profile {
	p := profiles[name]
	p.Profile = PowerPackProfile
	p.Regions = append([]Region(nil), p.Regions...)
	p.Pictures = append([]Picture(nil), p.Pictures...)
	l, ok := powerpackLayouts[name]
	if !ok {
		return p
	} // TABLE3/4 share the canonical linked layout.
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
	}
	return p
}

func translateLinked(name string, data []byte) ([]byte, error) {
	l, ok := powerpackLayouts[name]
	if !ok {
		return data, nil
	}
	return translateRecords(l, data)
}

func translateRecords(l linkedLayout, data []byte) ([]byte, error) {
	decoded := make([]byte, l.DecodedSize)
	for _, c := range l.Copies {
		if c.Source < 0 || c.Size < 0 || c.Source > len(data) || c.Size > len(data)-c.Source ||
			c.Destination < 0 || c.Destination > len(decoded) || c.Size > len(decoded)-c.Destination {
			return nil, fmt.Errorf("bounded linked record at %#x", c.Source)
		}
		copy(decoded[c.Destination:c.Destination+c.Size], data[c.Source:c.Source+c.Size])
	}
	// Copying preserves the actual input semantics. No canonical cue is injected.
	for _, j := range l.Jingles {
		got := tablelogic.JingleSpec{Position: decoded[j.Destination], Repeat: decoded[j.Destination+1], Priority: decoded[j.Destination+2]}
		want := tablelogic.JingleSpec{Position: j.Position, Repeat: j.Repeat, Priority: j.Priority}
		if got != want {
			return nil, fmt.Errorf("incompatible jingle %s", j.Role)
		}
	}
	return decoded, nil
}

// ValidateDecoded validates the shared decoder address contract, including
// registered edition semantics. Validate remains the strict canonical layout
// validator; source detection never uses this decoded-data entry point.
func ValidateDecoded(name string, data []byte) error {
	if _, ok := profiles[name]; !ok {
		return fmt.Errorf("%s: no supported data layout", name)
	}
	if err := Validate(name, data); err == nil {
		return nil
	}
	if err := validateProfile(name, data, linkedProfile(name, false)); err == nil {
		return nil
	}
	if err := validateProfile(name, data, deluxeDecodedProfile(name)); err == nil {
		return nil
	}
	if name == "TABLE1.PRG" {
		if err := validateProfile(name, data, partyLandDemoProfile(false)); err == nil {
			return nil
		}
	}
	return validateProfile(name, data, deluxeAltDecodedProfile(name))
}
