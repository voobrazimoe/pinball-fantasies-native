package datalayout

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestPowerPackDescriptorCoverage(t *testing.T) {
	for name, layout := range powerpackLayouts {
		last := 0
		for _, c := range layout.Copies {
			if c.Destination < last || c.Source < 0 || c.Size < 1 || c.Destination+c.Size > layout.DecodedSize {
				t.Fatal(name, c)
			}
			last = c.Destination + c.Size
		}
		for _, r := range profiles[name].Regions {
			layout.sourceOffset(r.Offset, r.Size)
		}
		for _, p := range profiles[name].Pictures {
			layout.sourceOffset(p.Offset, 12)
		}
		p := linkedProfile(name, true)
		if validateProfile(name, nil, p) == nil {
			t.Fatal("empty accepted")
		}
	}
	for _, name := range []string{"unknown", "INTRO.PRG", "TABLE1.PRG"} {
		if _, err := PreparePRG(name, []byte{0}); err == nil {
			t.Fatal("malformed accepted")
		}
	}
	if ValidateDecoded("unknown", nil) == nil {
		t.Fatal("unknown decoded input accepted")
	}
	if _, err := DetectInstallation(nil); err == nil {
		t.Fatal("incomplete accepted")
	}
}

func TestPrivatePowerPackMappingsAndCoherence(t *testing.T) {
	a, b := os.Getenv("PF_RUNTIME_DATA"), os.Getenv("PF_POWERPACK_DATA")
	if a == "" || b == "" {
		t.Skip("set PF_RUNTIME_DATA and PF_POWERPACK_DATA to private installations")
	}
	read := func(dir string) map[string][]byte {
		out := map[string][]byte{}
		for _, n := range prgNames {
			raw, err := os.ReadFile(filepath.Join(dir, n))
			if err != nil {
				t.Fatal(err)
			}
			out[n] = raw
		}
		return out
	}
	aa, bb := read(a), read(b)
	for _, tc := range []struct {
		inputs map[string][]byte
		want   string
	}{{aa, RetailProfile}, {bb, PowerPackProfile}} {
		id, err := DetectInstallation(tc.inputs)
		if err != nil || id != tc.want {
			t.Fatal(id, err)
		}
	}
	for _, name := range prgNames {
		got, err := PreparePRGForProfile(PowerPackProfile, name, bb[name])
		if err != nil {
			t.Fatal(name, err)
		}
		for _, r := range profiles[name].Regions {
			expected := aa[name][r.Offset : r.Offset+r.Size]
			if name == "TABLE1.PRG" && r.Purpose == "matrix record S_EMPTY" {
				expected = []byte{62, 0, 0}
			}
			if !bytes.Equal(got[r.Offset:r.Offset+r.Size], expected) {
				t.Fatal(name, r)
			}
		}
		for _, p := range profiles[name].Pictures {
			// Full FORM extents are reviewed in the descriptor, not only headers.
			size := 8 + int(binary.BigEndian.Uint32(aa[name][p.Offset+4:p.Offset+8]))
			if !bytes.Equal(got[p.Offset:p.Offset+size], aa[name][p.Offset:p.Offset+size]) {
				t.Fatal(name, p)
			}
		}
		if name == "TABLE3.PRG" || name == "TABLE4.PRG" {
			if !bytes.Equal(got, aa[name]) || &got[0] != &bb[name][0] {
				t.Fatal("shared canonical path changed", name)
			}
		}
		if _, ok := powerpackLayouts[name]; ok {
			bad := append([]byte(nil), bb[name]...)
			bad[linkedProfile(name, true).Pictures[0].Offset] ^= 1
			if _, err = PreparePRGForProfile(PowerPackProfile, name, bad); err == nil {
				t.Fatal("malformed accepted")
			}
			for _, r := range linkedProfile(name, true).Regions {
				if r.SHA256 != "" {
					bad = append([]byte(nil), bb[name]...)
					bad[r.Offset] ^= 1
					if _, err = PreparePRGForProfile(PowerPackProfile, name, bad); err == nil {
						t.Fatal("damaged record accepted", name, r)
					}
					break
				}
			}
		}
	}
	// Reject priority normalization in either source profile, even though both
	// priorities have registered meanings in the shared decoded address contract.
	for _, tc := range []struct {
		data    []byte
		profile string
		at      int
	}{{aa["TABLE1.PRG"], RetailProfile, 0x1a9ae}, {bb["TABLE1.PRG"], PowerPackProfile, 0x1a9be}} {
		bad := append([]byte(nil), tc.data...)
		bad[tc.at] ^= 1
		if _, err := PreparePRGForProfile(tc.profile, "TABLE1.PRG", bad); err == nil {
			t.Fatal("source cue semantic normalization accepted", tc.profile)
		}
	}
	// Exhaust all six hybrids across the three edition-distinguishing inputs.
	for mask := 1; mask < 7; mask++ {
		hybrid := read(a)
		for i, n := range prgNames[:3] {
			if mask&(1<<i) != 0 {
				hybrid[n] = bb[n]
			}
		}
		if _, err := DetectInstallation(hybrid); err == nil {
			t.Fatal("hybrid accepted", mask)
		}
	}
}
