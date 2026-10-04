package datalayout

import "testing"

func TestProfilesExcludeUnusedExecutableHeader(t *testing.T) {
	for name, p := range profiles {
		if len(p.Pictures) == 0 {
			t.Fatal(name, "missing structural artwork anchors")
		}
		for _, r := range p.Regions {
			if r.Offset <= 0 || r.Size <= 0 {
				t.Fatal(name, "header byte zero must never be consumed", r)
			}
			if r.SHA256 != "" && len(r.SHA256) != 64 {
				t.Fatal(name, "bad regional fingerprint")
			}
		}
		for _, pic := range p.Pictures {
			if pic.Offset <= 0 || pic.Width < 1 || pic.Height < 1 || pic.Planes < 1 || pic.Planes > 8 {
				t.Fatal(name, "invalid picture geometry")
			}
		}
		if Validate(name, nil) == nil {
			t.Fatal(name, "truncation accepted")
		}
	}
}
