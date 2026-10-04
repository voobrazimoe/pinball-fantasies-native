package frontend

import (
	"fmt"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/gameshow"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/speeddevils"
	"pinballfantasies/internal/stones"
	"testing"
)

func TestFourTablesUseOneDOSMouseAndKeyboardSpring(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			r := configured(t, settings.Legacy())
			runKey(t, r, Space)
			runKey(t, r, Key(int(F1)+table-1))
			runKey(t, r, F1)
			runTicks(t, r, 200)
			var p *physics.Game
			switch g := r.Model.Session.(type) {
			case *partyland.Game:
				p = g.Physics
			case *speeddevils.Game:
				p = g.Physics
			case *gameshow.Game:
				p = g.Physics
			case *stones.Game:
				p = g.Physics
			default:
				t.Fatalf("missing table %d", table)
			}
			sync := func(in gameplay.Controls) {
				t.Helper()
				if err := r.Update(Input{Gameplay: in}); err != nil {
					t.Fatal(err)
				}
			}
			for i := 0; i < 4; i++ {
				sync(gameplay.Controls{Down: true})
			}
			if p.SpringPosition != 4 {
				t.Fatal("keyboard charge", p.SpringPosition)
			}
			sync(gameplay.Controls{MouseY: 1})
			if p.SpringPosition != 5 {
				t.Fatal("mouse down", p.SpringPosition)
			}
			sync(gameplay.Controls{MouseY: -1})
			if p.SpringPosition != 4 {
				t.Fatal("mouse up", p.SpringPosition)
			}
			for i := 0; i < 40; i++ {
				sync(gameplay.Controls{MouseY: -1})
			}
			if p.SpringPosition != 0 {
				t.Fatal("mouse lower clamp")
			}
			for i := 0; i < 40; i++ {
				sync(gameplay.Controls{MouseY: 1})
			}
			if p.SpringPosition != 32 {
				t.Fatal("mouse upper clamp")
			}
			sync(gameplay.Controls{MouseFire: true})
			if p.SpringPosition != 0 || p.Ball.VY >= 0 {
				t.Fatal("mouse launch/reset", p.SpringPosition, p.Ball.VY)
			}
		})
	}
}

// Original-backed integration: the additive touch edge reaches all four real
// spring tasks and uses their existing release callbacks/physics.
func TestFourTablesUseAbsoluteTouchSpring(t *testing.T) {
	for table := 1; table <= 4; table++ {
		for _, target := range []int{16, 32} {
			t.Run(fmt.Sprintf("%d/%d", table, target), func(t *testing.T) {
				r := configured(t, settings.Legacy())
				runKey(t, r, Space)
				runKey(t, r, Key(int(F1)+table-1))
				runKey(t, r, F1)
				runTicks(t, r, 200)
				var p *physics.Game
				switch g := r.Model.Session.(type) {
				case *partyland.Game:
					p = g.Physics
				case *speeddevils.Game:
					p = g.Physics
				case *gameshow.Game:
					p = g.Physics
				case *stones.Game:
					p = g.Physics
				default:
					t.Fatal("missing table")
				}
				if err := r.Update(Input{Gameplay: gameplay.Controls{TouchSet: true, TouchTarget: target}}); err != nil {
					t.Fatal(err)
				}
				if int(p.SpringPosition) != target {
					t.Fatal("touch charge", p.SpringPosition, target)
				}
				if err := r.Update(Input{Gameplay: gameplay.Controls{MouseFire: true}}); err != nil {
					t.Fatal(err)
				}
				if p.SpringPosition != 0 || p.Ball.VY >= 0 {
					t.Fatal("touch release physics", p.SpringPosition, p.Ball.VY)
				}
			})
		}
	}
}
