package datalayout

import (
	"os"
	"path/filepath"
	"testing"
)

// Inputs are read only and optional in public CI; no commercial fixtures stored.
func TestPossessedCanonicalAndNearMisses(t *testing.T) {
	dir := os.Getenv("PF_RUNTIME_DATA")
	if dir == "" {
		t.Skip("set PF_RUNTIME_DATA to canonical possessed inputs")
	}
	for name, p := range profiles {
		t.Run(name, func(t *testing.T) {
			b, e := os.ReadFile(filepath.Join(dir, name))
			if e != nil {
				t.Fatal(e)
			}
			if _, e = PreparePRG(name, b); e != nil {
				t.Fatal(e)
			}
			for _, pic := range p.Pictures {
				bad := append([]byte(nil), b...)
				bad[pic.Offset] ^= 1
				if _, e = PreparePRG(name, bad); e == nil {
					t.Fatal("malformed IFF anchor accepted")
				}
			}
			for _, r := range p.Regions {
				if r.SHA256 == "" {
					continue
				}
				bad := append([]byte(nil), b...)
				bad[r.Offset] ^= 1
				if _, e = PreparePRG(name, bad); e == nil {
					t.Fatal("changed semantic record accepted", r.Purpose)
				}
			}
		})
	}
}
func TestPossessedPowerPackIndependently(t *testing.T) {
	dir := os.Getenv("PF_POWERPACK_DATA")
	if dir == "" {
		t.Skip("set PF_POWERPACK_DATA to audited alternate inputs")
	}
	for name := range profiles {
		b, e := os.ReadFile(filepath.Join(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		_, e = PreparePRG(name, b)
		want := true
		if (e == nil) != want {
			t.Fatalf("%s acceptance=%v want %v", name, e == nil, want)
		}
	}
}
