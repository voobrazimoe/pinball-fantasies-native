package datalayout

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestPrivateDeluxeAltMappingsAndSeeds(t *testing.T) {
	c, d := privatePRGs(t, "PF_DELUXE_CD_ALT_DATA"), privatePRGs(t, "PF_DELUXE_CD_DATA")
	id, err := DetectInstallation(c)
	if err != nil || id != DeluxeCDAltProfile {
		t.Fatal(id, err)
	}
	for _, name := range prgNames {
		t.Run(name, func(t *testing.T) {
			l, _ := deluxeAltLayout(name)
			out, err := PreparePRGForProfile(id, name, c[name])
			if err != nil {
				t.Fatal(err)
			}
			other, err := PreparePRGForProfile(DeluxeCDProfile, name, d[name])
			if err != nil {
				t.Fatal(err)
			}
			for _, span := range l.Copies {
				if !bytes.Equal(out[span.Destination:span.Destination+span.Size], c[name][span.Source:span.Source+span.Size]) {
					t.Fatal(span)
				}
			}
			for _, pic := range l.SourceProfile.Pictures {
				bad := bytes.Clone(c[name])
				bad[pic.Offset] ^= 1
				if _, e := PreparePRGForProfile(id, name, bad); e == nil {
					t.Fatal("bad picture", pic)
				}
			}
			if name == "INTRO.PRG" {
				if !bytes.Equal(out, other) {
					t.Fatal("shared presentation differs")
				}
				layout, e := DecodedFrontendLayout(out)
				if e != nil || len(layout.Pictures) != 17 || layout.StartupLowerY != 123 || layout.Pictures[7].Height != 123 || layout.Pictures[8].Height != 117 {
					t.Fatal(layout, e)
				}
				return
			}
			if len(l.Copies) != len(deluxeLayouts[name].Copies) {
				t.Fatal("TABLE mapping not shared")
			}
			seed := deluxeAltMetadata.Seeds[name]
			initials, e := FactoryInitials(name, out)
			if e != nil {
				t.Fatal(e)
			}
			for rank, want := range seed.Initials {
				if string(initials[rank][:]) != want {
					t.Fatal(initials)
				}
				for ch := 0; ch < 3; ch++ {
					for _, tc := range []struct {
						src         []byte
						profile     string
						replacement byte
					}{{c[name], id, d[name][seed.Source+rank*16+12+ch]}, {d[name], DeluxeCDProfile, c[name][seed.Source+rank*16+12+ch]}} {
						bad := bytes.Clone(tc.src)
						bad[seed.Source+rank*16+12+ch] = tc.replacement
						if _, e := PreparePRGForProfile(tc.profile, name, bad); e == nil {
							t.Fatal("partial seed accepted", rank, ch, tc.profile)
						}
					}
				}
				// Each whole entry from the other family must fail too.
				bad := bytes.Clone(c[name])
				copy(bad[seed.Source+rank*16+12:], d[name][seed.Source+rank*16+12:seed.Source+rank*16+15])
				if _, e := PreparePRGForProfile(id, name, bad); e == nil {
					t.Fatal("D entry in C")
				}
			}
			// Every proper subset of the four factory entries is a forbidden hybrid.
			for mask := 1; mask < 15; mask++ {
				for _, tc := range []struct {
					base, other []byte
					profile     string
				}{{c[name], d[name], id}, {d[name], c[name], DeluxeCDProfile}} {
					bad := bytes.Clone(tc.base)
					for rank := 0; rank < 4; rank++ {
						if mask&(1<<rank) != 0 {
							at := seed.Source + rank*16 + 12
							copy(bad[at:at+3], tc.other[at:at+3])
						}
					}
					if _, e := PreparePRGForProfile(tc.profile, name, bad); e == nil {
						t.Fatal("seed hybrid", mask)
					}
				}
			}
			// Only the twelve reviewed initials differ in the prepared contract.
			normalized := bytes.Clone(out)
			for rank := 0; rank < 4; rank++ {
				at := seed.Destination + rank*16 + 12
				copy(normalized[at:at+3], other[at:at+3])
			}
			if !bytes.Equal(normalized, other) {
				t.Fatal("unreviewed decoded difference")
			}
			for _, s := range l.Selectors {
				if binary.LittleEndian.Uint16(c[name][s.Source:]) != s.Raw || binary.LittleEndian.Uint16(out[s.Destination:]) != s.NativeID {
					t.Fatal(s)
				}
				bad := bytes.Clone(c[name])
				binary.LittleEndian.PutUint16(bad[s.Source:], 0xffff)
				if _, e := PreparePRGForProfile(id, name, bad); e == nil {
					t.Fatal("unknown selector")
				}
				bad = bytes.Clone(out)
				binary.LittleEndian.PutUint16(bad[s.Destination:], s.Raw)
				if ValidateDecoded(name, bad) == nil {
					t.Fatal("raw selector accepted")
				}
			}
			for _, j := range l.Jingles {
				if !bytes.Equal(c[name][j.Source:j.Source+3], out[j.Destination:j.Destination+3]) {
					t.Fatal("cue not preserved")
				}
			}
		})
	}
}

func TestPrivateFourProfileHybridMatrix(t *testing.T) {
	sets := []map[string][]byte{privatePRGs(t, "PF_RUNTIME_DATA"), privatePRGs(t, "PF_POWERPACK_DATA"), privatePRGs(t, "PF_DELUXE_CD_ALT_DATA"), privatePRGs(t, "PF_DELUXE_CD_DATA")}
	ids := []string{RetailProfile, PowerPackProfile, DeluxeCDAltProfile, DeluxeCDProfile}
	for i, s := range sets {
		if id, e := DetectInstallation(s); e != nil || id != ids[i] {
			t.Fatal(id, e)
		}
	}
	// All 4^5 role combinations, including INTRO/table family crossings and pairs.
	// A/B TABLE3/4 are identical, so accept combinations only if their actual
	// bytes constitute one of the four coherent installations.
	for code := 0; code < 1024; code++ {
		mix := map[string][]byte{}
		n := code
		for _, role := range prgNames {
			mix[role] = sets[n%4][role]
			n /= 4
		}
		expected := ""
		for i, s := range sets {
			same := true
			for _, role := range prgNames {
				if !bytes.Equal(mix[role], s[role]) {
					same = false
				}
			}
			if same {
				expected = ids[i]
			}
		}
		id, e := DetectInstallation(mix)
		if expected == "" {
			if e == nil {
				t.Fatal("hybrid accepted", fmt.Sprintf("%05b", code), id)
			}
		} else if e != nil || id != expected {
			t.Fatal(code, id, e)
		}
	}
}
