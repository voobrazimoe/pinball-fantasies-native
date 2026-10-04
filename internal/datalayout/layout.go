// Package datalayout identifies supported consumed-data layouts, independently
// of whole-file inventory identity. It never loads or executes DOS code.
package datalayout

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed profiles.json
var profileJSON []byte

type Region struct {
	Offset, Size int
	Purpose      string
	SHA256       string
	Kind         string
}
type Picture struct {
	Offset, Width, Height, Planes int
	Kind                          string
}
type Profile struct {
	Profile  string
	Regions  []Region
	Pictures []Picture
}

var profiles = func() map[string]Profile {
	var p map[string]Profile
	if err := json.Unmarshal(profileJSON, &p); err != nil {
		panic(err)
	}
	return p
}()

// Validate checks all non-IFF reads before any fixed-offset decoder runs.
// Exact region hashes protect control/physics data coupled to native source
// semantics. Opaque artwork needs only bounds; IFFs are checked by assets.
func Validate(name string, data []byte) error {
	p, ok := profiles[name]
	if !ok {
		return fmt.Errorf("%s: no supported data layout", name)
	}
	for _, picture := range p.Pictures {
		if err := validatePicture(data, picture); err != nil {
			return fmt.Errorf("%s: unsupported layout %s: %w", name, p.Profile, err)
		}
	}
	for _, r := range p.Regions {
		if r.Offset < 0 || r.Size < 0 || r.Offset > len(data) || r.Size > len(data)-r.Offset {
			return fmt.Errorf("%s: unsupported layout %s: truncated %s at %#x", name, p.Profile, r.Purpose, r.Offset)
		}
		raw := data[r.Offset : r.Offset+r.Size]
		if r.Kind == "scroll" {
			if len(raw) < 21 {
				return fmt.Errorf("%s: unsupported layout: short scroll record", name)
			}
			for _, ch := range raw {
				if ch != 1 && ch != 255 && (ch < 32 || ch > 96) {
					return fmt.Errorf("%s: unsupported layout: scroll character outside glyph map", name)
				}
			}
		}
		if r.Kind == "glyph" {
			if len(raw)%4 != 1 || raw[len(raw)-1] != 0xc3 {
				return fmt.Errorf("%s: unsupported layout: glyph terminator", name)
			}
			for q := 0; q < len(raw)-1; q += 4 {
				if raw[q] != 0x88 || (raw[q+1] != 0x87 && raw[q+1] != 0xa7) {
					return fmt.Errorf("%s: unsupported layout: literal glyph record", name)
				}
			}
		}

		if r.SHA256 != "" && fmt.Sprintf("%x", sha256.Sum256(data[r.Offset:r.Offset+r.Size])) != r.SHA256 {
			return fmt.Errorf("%s: unsupported layout %s: incompatible %s at %#x", name, p.Profile, r.Purpose, r.Offset)
		}
	}
	return nil
}

func Pictures(name string) []Picture { return append([]Picture(nil), profiles[name].Pictures...) }
