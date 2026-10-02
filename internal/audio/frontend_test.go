package audio

import (
	"bytes"
	"os"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func TestFrontendModules(t *testing.T) {
	for _, name := range []string{"INTRO.MOD", "MOD2.MOD"} {
		t.Run(name, func(t *testing.T) {
			testinputs.Require(t, "../../"+name)
			b, e := os.ReadFile("../../" + name)
			if e != nil {
				t.Fatal(e)
			}
			m, e := DecodeFrontend(b)
			if e != nil {
				t.Fatal(e)
			}
			a, c := New(m), New(m)
			var pcm []byte
			for i := 0; i < 2400; i++ {
				p, q := a.Render(Rate/60), c.Render(Rate/60)
				if !bytes.Equal(p, q) {
					t.Fatal("nondeterministic frontend tracker")
				}
				pcm = append(pcm, p...)
			}
			if bytes.Count(pcm, []byte{0}) == len(pcm) {
				t.Fatal("frontend music is silent")
			}
			if name == "INTRO.MOD" {
				for _, s := range m.Samples[20:] {
					if len(s.PCM) != 2 || s.PCM[0] != 0 || s.PCM[1] != 0 {
						t.Fatal("omitted silent slot not normalized")
					}
				}
			}
			b[0] ^= 1
			if _, e := DecodeFrontend(b); e == nil {
				t.Fatal("unknown module accepted")
			}
		})
	}
}
