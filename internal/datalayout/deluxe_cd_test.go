package datalayout

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestDeluxeDescriptorCoverage(t *testing.T) {
	for name, l := range deluxeLayouts {
		last := 0
		for _, c := range l.Copies {
			if c.Destination < last || c.Size <= 0 || c.Source < 0 || c.Destination+c.Size > l.DecodedSize {
				t.Fatal(name, c)
			}
			last = c.Destination + c.Size
			for _, s := range l.Selectors {
				if s.Destination < c.Destination+c.Size && s.Destination+2 > c.Destination {
					t.Fatal("typed selector included in byte copy", s)
				}
			}
		}
		for _, r := range profiles[name].Regions {
			if r.Purpose == "area handlers/rectangles/terminator" {
				continue
			}
			size := r.Size
			if r.Purpose == "tower packed bitmap" {
				size = 6680
			}
			l.sourceOffset(r.Offset, size)
		}
	}
	l := deluxeLayouts["TABLE4.PRG"]
	if len(l.Selectors) != 44 {
		t.Fatal(len(l.Selectors))
	}
	roles := map[string]bool{}
	for _, s := range l.Selectors {
		roles[s.Role] = true
	}
	if len(roles) != 34 {
		t.Fatal(len(roles))
	}
}

func privatePRGs(t *testing.T, env string) map[string][]byte {
	t.Helper()
	dir := os.Getenv(env)
	if dir == "" {
		t.Skip("set " + env + " to private installation")
	}
	out := map[string][]byte{}
	for _, name := range prgNames {
		b, e := os.ReadFile(filepath.Join(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		out[name] = b
	}
	return out
}
func TestPrivateDeluxeMappings(t *testing.T) {
	a, d := privatePRGs(t, "PF_RUNTIME_DATA"), privatePRGs(t, "PF_DELUXE_CD_DATA")
	id, e := DetectInstallation(d)
	if e != nil || id != DeluxeCDProfile {
		t.Fatal(id, e)
	}
	for _, name := range prgNames {
		out, e := PreparePRGForProfile(id, name, d[name])
		if e != nil {
			t.Fatal(name, e)
		}
		l := deluxeLayouts[name]
		for _, picture := range l.SourceProfile.Pictures {
			bad := append([]byte(nil), d[name]...)
			bad[picture.Offset] ^= 1
			if _, err := PreparePRGForProfile(id, name, bad); err == nil {
				t.Fatal("malformed picture accepted", name, picture)
			}
		}
		for _, c := range l.Copies {
			if !bytes.Equal(out[c.Destination:c.Destination+c.Size], d[name][c.Source:c.Source+c.Size]) {
				t.Fatal(name, c)
			}
		}
		for _, r := range profiles[name].Regions {
			size := r.Size
			if r.Purpose == "tower packed bitmap" {
				size = 6680
			}
			want := a[name][r.Offset : r.Offset+size]
			if name == "TABLE1.PRG" && r.Purpose == "matrix record S_EMPTY" {
				want = []byte{62, 0, 0}
			}
			if !bytes.Equal(out[r.Offset:r.Offset+size], want) {
				t.Fatal("decoded region", name, r)
			}
		}
		for _, s := range l.Selectors {
			if binary.LittleEndian.Uint16(out[s.Destination:]) != s.NativeID {
				t.Fatal(s)
			}
			// Reverting even one word to numeric DOS representation must fail.
			naive := append([]byte(nil), out...)
			binary.LittleEndian.PutUint16(naive[s.Destination:], s.Raw)
			if ValidateDecoded(name, naive) == nil {
				t.Fatal("naive selector accepted", s)
			}
			bad := append([]byte(nil), d[name]...)
			binary.LittleEndian.PutUint16(bad[s.Source:], 0xffff)
			if _, e := PreparePRGForProfile(id, name, bad); e == nil {
				t.Fatal("unknown selector accepted", s)
			}
		}
		if name == "INTRO.PRG" {
			layout, e := DecodedFrontendLayout(out)
			if e != nil {
				t.Fatal(e)
			}
			if len(layout.Pictures) != 17 || layout.Pictures[7].Height != 123 || layout.Pictures[8].Height != 117 || layout.StartupLowerY != 123 {
				t.Fatal(layout)
			}
			for i, p := range l.SourceProfile.Pictures {
				got := layout.Pictures[i]
				if p.Width != got.Width || p.Height != got.Height || p.Planes != got.Planes || p.Kind != got.Kind {
					t.Fatal(i, p, got)
				}
			}
		}
	}
}
func TestPrivateDeluxeCoherenceAndCRejection(t *testing.T) {
	a, b, d := privatePRGs(t, "PF_RUNTIME_DATA"), privatePRGs(t, "PF_POWERPACK_DATA"), privatePRGs(t, "PF_DELUXE_CD_DATA")
	for _, other := range []map[string][]byte{a, b} {
		// Every proper subset includes all single-role substitutions and both INTRO
		// directions, separated TABLE1/TABLE4, and the dangerous combined cases.
		for mask := 1; mask < 31; mask++ {
			mix := map[string][]byte{}
			for i, n := range prgNames {
				mix[n] = d[n]
				if mask&(1<<i) != 0 {
					mix[n] = other[n]
				}
			}
			if id, e := DetectInstallation(mix); e == nil {
				t.Fatal("hybrid accepted", mask, id)
			}
		}
	}
	t.Run("C", func(t *testing.T) {
		c := privatePRGs(t, "PF_UNSUPPORTED_CD_DATA")
		if id, e := DetectInstallation(c); e == nil {
			t.Fatal("C accepted", id)
		}
		for _, n := range prgNames {
			if _, e := PreparePRGForProfile(DeluxeCDProfile, n, c[n]); e == nil {
				t.Fatal("C role accepted as D", n)
			}
			mix := map[string][]byte{}
			for _, key := range prgNames {
				mix[key] = d[key]
			}
			mix[n] = c[n]
			if id, e := DetectInstallation(mix); e == nil {
				t.Fatal("D/C hybrid accepted", n, id)
			}
		}
	})
	for _, n := range prgNames {
		bad := map[string][]byte{}
		for _, key := range prgNames {
			bad[key] = d[key]
		}
		bad[n] = []byte{0}
		if _, e := DetectInstallation(bad); e == nil {
			t.Fatal("malformed accepted", n)
		}
		bad[n] = append([]byte{0}, d[n]...)
		if _, e := DetectInstallation(bad); e == nil {
			t.Fatal("unknown layout accepted", n)
		}
		delete(bad, n)
		if _, e := DetectInstallation(bad); e == nil {
			t.Fatal("incomplete accepted", n)
		}
	}
}
