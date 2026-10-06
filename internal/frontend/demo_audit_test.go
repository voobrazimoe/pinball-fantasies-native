package frontend

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/datalayout"
	"reflect"
	"testing"
)

// This research contract must not become a production acceptance path. DMO0
// still lacks a complete callback/control graph; keep demo fail-closed.
func TestPrivate10MinuteDemoResearchBoundary(t *testing.T) {
	dir, canonical := os.Getenv("PF_10MIN_DEMO_DATA"), os.Getenv("PF_RUNTIME_DATA")
	if dir == "" || canonical == "" {
		t.Skip("set PF_10MIN_DEMO_DATA and PF_RUNTIME_DATA for private DMO0 evidence")
	}
	names := []string{"INTRO.PRG", "INTRO.MOD", "MOD2.MOD", "TABLE1.PRG", "TABLE1.MOD"}
	want := []string{
		"05bdba35e0a9a31a87b944428ddad27a00ba8983e97963290ab2ee1f57910fa3",
		"f36beae00efec1dd9e1c4e977bea264b7ec41ab18f258528ab66577a9ec66613",
		"aa5003c275b494062f37f44e8c77105b8a420555f4bd6ff53d7698f89c540f21",
		"44b8f4b76ee16c47cda26904681e83e8f4420974bcea249d099790bfbd7369e3",
		"a0877e4372abe64b70d9e361bf257ea5a84c948771f0eace3433d5f6399060b5",
	}
	read := func(root, name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	inputs, identity := map[string][]byte{}, ""
	for i, n := range names {
		b := read(dir, n)
		if fmt.Sprintf("%x", sha256.Sum256(b)) != want[i] {
			t.Fatalf("%s is not the pinned DMO0 fixture", n)
		}
		inputs[n] = b
		identity += n + "\x00" + want[i] + "\n"
	}
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(identity))); got != "e7d9aaedf4f06f67d1553d88be0b0f87568bb1da49344b1a7f070f37c160809e" {
		t.Fatal(got)
	}
	for _, tc := range []struct {
		name   string
		decode func([]byte) (*audio.Module, error)
	}{{"INTRO.MOD", audio.DecodeIntro}, {"MOD2.MOD", audio.DecodeMenu}, {"TABLE1.MOD", audio.Decode}} {
		x, e := tc.decode(inputs[tc.name])
		if e != nil {
			t.Fatal(tc.name, e)
		}
		y, e := tc.decode(read(canonical, tc.name))
		if e != nil || !reflect.DeepEqual(x, y) {
			t.Fatal(tc.name, "decoded audio differs", e)
		}
	}
	for _, n := range []string{"INTRO.PRG", "TABLE1.PRG"} {
		if _, e := datalayout.PreparePRG(n, inputs[n]); e == nil {
			t.Fatal("unreviewed demo accepted", n)
		}
	}
	if _, e := Load(dir, nil); e == nil {
		t.Fatal("five-file demo booted through full loader")
	}
	roles := []string{"INTRO.PRG", "TABLE1.PRG", "TABLE2.PRG", "TABLE3.PRG", "TABLE4.PRG"}
	for _, env := range []string{"PF_RUNTIME_DATA", "PF_POWERPACK_DATA", "PF_DELUXE_CD_ALT_DATA", "PF_DELUXE_CD_DATA"} {
		t.Run(env, func(t *testing.T) {
			root := os.Getenv(env)
			if root == "" {
				t.Skip("supply " + env)
			}
			full := map[string][]byte{}
			for _, n := range roles {
				full[n] = read(root, n)
			}
			if _, e := datalayout.DetectInstallation(full); e != nil {
				t.Fatal("full baseline rejected", e)
			}
			for _, n := range []string{"INTRO.PRG", "TABLE1.PRG"} {
				original := full[n]
				full[n] = inputs[n]
				if _, e := datalayout.DetectInstallation(full); e == nil {
					t.Fatal("demo/full hybrid accepted", n)
				}
				full[n] = original
			}
			full["INTRO.PRG"], full["TABLE1.PRG"] = inputs["INTRO.PRG"], inputs["TABLE1.PRG"]
			if _, e := datalayout.DetectInstallation(full); e == nil {
				t.Fatal("demo plus unrelated TABLE2-4 accepted")
			}
		})
	}
}
