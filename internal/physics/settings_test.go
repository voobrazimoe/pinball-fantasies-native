package physics

import (
	"encoding/json"
	"fmt"
	"os"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func TestSourceAngleAllTables(t *testing.T) {
	decoders := []func([]byte) (*Table, error){DecodePartyLand, DecodeSpeedDevils, DecodeGameshow, DecodeStones}
	ramps := [][][2]int16{{{0, 10}, {2, 14}, {-2, 14}, {-4, 16}}, {{0, 10}, {0, 15}, {0, 25}, {-1, 10}, {0, 20}, {12, 10}}, {{0, 10}, {4, 12}, {0, 14}, {2, 9}, {6, 13}}, {{0, 10}, {-10, 5}, {0, -10}, {5, 0}, {5, 15}, {-10, 12}, {2, 15}, {-8, 12}, {3, 10}, {4, 13}, {7, 10}}}
	for i, decode := range decoders {
		testinputs.Require(t, fmt.Sprintf("../../TABLE%d.PRG", i+1))
		b, e := os.ReadFile(fmt.Sprintf("../../TABLE%d.PRG", i+1))
		if e != nil {
			t.Fatal(e)
		}
		table, e := decode(b)
		if e != nil {
			t.Fatal(e)
		}
		for angle := byte(0); angle < 2; angle++ {
			c := settings.Legacy()
			c.Angle = angle
			g := New(table)
			g.Configure(c, ramps[i])
			for n, v := range ramps[i] {
				want := v[1] - 3*int16(angle)
				if g.gravity[n] != [2]int16{v[0], want} {
					t.Fatal(i, n, g.gravity[n])
				}
			}
			// Controlled free-ball arithmetic from first SC_PROGRAM through a sync;
			// bypass collision deliberately while retaining signed integer integration.
			clean := *g.Table
			clean.mask11 = make([]byte, 23040)
			g.Table = &clean
			g.mask12 = make([]byte, 23040)
			g.SetBall(100, 100, 0, 0, false)
			g.Ball.GX, g.Ball.GY = g.gravity[0][0], g.gravity[0][1]
			for tick := 0; tick < 3; tick++ {
				if e = g.step(Inputs{}); e != nil {
					t.Fatal(e)
				}
			}
			gravity := int32(10 - 3*angle)
			if g.Ball.Y != 100*1024+3*gravity || g.Ball.VY != int16(3*gravity) {
				t.Fatal("first sync", i, angle, g.Ball)
			}
			for tick := 0; tick < 9; tick++ {
				if e = g.step(Inputs{}); e != nil {
					t.Fatal(e)
				}
			}
			if g.Ball.Y != 100*1024+66*gravity || g.Ball.VY != int16(12*gravity) {
				t.Fatal("controlled ball", i, angle, g.Ball)
			}
		}
	}
}
func TestSourceScrollValuesAndForcedRaster(t *testing.T) {
	var fixture struct {
		Scroll []struct {
			Raster []int16 `json:"raster"`
		} `json:"scroll"`
	}
	raw, e := os.ReadFile("../../analysis/pf11-settings-fixtures.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(raw, &fixture); e != nil {
		t.Fatal(e)
	}

	for resolution := byte(0); resolution < 2; resolution++ {
		for value := byte(0); value < 3; value++ {
			for index, y := range []int16{0, 100, 200, 450, 570} {
				c := settings.Defaults()
				c.Resolution = resolution
				c.ScrollMode = settings.ScrollMode(value)
				g := New(table(t))
				g.Settings = c
				g.Raster = 1600
				g.Ball.PixelY = y
				expected := fixture.Scroll[int(resolution)*3+int(value)].Raster[index]
				before := g.Ball
				g.scroll()
				if g.Raster != expected || g.Ball != before || g.Syncs != 0 {
					t.Fatal(resolution, value, y, g.Raster, expected)
				}
				g.TargetRaster = 187
				g.Raster = 1600
				g.scroll()
				if g.Raster == (187+33)*16 {
					t.Fatal("SCREENFORCE2 bypassed smoothing")
				}
				g.ScrollForce = func() int16 { return 187 }
				g.scroll()
				if g.Raster != (187+33)*16 {
					t.Fatal("force lost")
				}
			}
		}
	}
}
func TestSourceMonoLampRounding(t *testing.T) {
	g := New(table(t))
	g.ReferenceMode = 1
	on := g.LampRGB([]byte{95, 70, 27}, true)
	off := g.LampRGB([]byte{95, 70, 27}, false)
	if on[0] != 162 || on[1] != 162 || on[2] != 162 || off[0] != 81 || off[1] != 81 || off[2] != 81 {
		t.Fatal(on, off)
	}
}
