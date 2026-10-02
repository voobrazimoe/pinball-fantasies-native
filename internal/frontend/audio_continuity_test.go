package frontend

import (
	"bytes"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"testing"
)

func TestSelectorAudioContinuity(t *testing.T) {
	for _, menu := range []bool{false, true} {
		t.Run(map[bool]string{false: "first-intro", true: "return-menu"}[menu], func(t *testing.T) {
			testinputs.Require(t, "../../INTRO.PRG", "../../INTRO.MOD", "../../MOD2.MOD", "../../TABLE1.PRG", "../../TABLE1.MOD", "../../TABLE2.PRG", "../../TABLE2.MOD", "../../TABLE3.PRG", "../../TABLE3.MOD", "../../TABLE4.PRG", "../../TABLE4.MOD", "../../PINBALL.CFG")
			r, e := Load("../..", nil)
			if e != nil {
				t.Fatal(e)
			}
			if e = r.Update(Input{Keys: []Key{Space}}); e != nil {
				t.Fatal(e)
			}
			if menu {
				r.Player = audio.New(r.Menu)
			}
			source := r.AudioSource()
			control := *r.Player
			transitions := 0
			for tick := 0; tick < 4000; tick++ {
				in := Input{}
				// Exercise both automatic transitions and repeated make edges.
				if tick%271 == 0 {
					in.Keys = []Key{Space}
				}
				before := r.Model.Mode
				if e = r.Update(in); e != nil {
					t.Fatal(e)
				}
				if before != r.Model.Mode {
					transitions++
				}
				expected := control.Render(audio.Rate / 60)
				if r.AudioSource() != source || !bytes.Equal(r.PCM, expected) || !reflect.DeepEqual(*r.Player, control) {
					t.Fatalf("audio continuity lost at tick%d mode %s -> %s order/row/tick %d/%d/%d", tick, before, r.Model.Mode, r.Player.Order, r.Player.Row, r.Player.Tick)
				}
			}
			if transitions < 10 {
				t.Fatal("insufficient selector/text transitions")
			}
		})
	}
}
