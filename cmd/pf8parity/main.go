// pf8parity exports native logical frames. It never invokes a DOS runtime.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/speeddevils"
)

func main() {
	dir := "analysis/pf8-checkpoints"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	must(os.MkdirAll(dir, 0755))
	r, e := frontend.Load(".", nil)
	must(e)
	manifest := map[string]any{}
	save := func(name string, im *image.RGBA) {
		p := filepath.Join(dir, name+".png")
		f, e := os.Create(p)
		must(e)
		must(png.Encode(f, im))
		must(f.Close())
		manifest[name] = map[string]any{"dimensions": im.Rect.Size(), "rgba_sha256": fmt.Sprintf("%x", sha256.Sum256(im.Pix))}
	}
	sync := func(in frontend.Input) { must(r.Update(in)) }
	for i := 0; i < 100; i++ {
		sync(frontend.Input{})
	}
	save("01-griffin", r.Frame())
	sync(frontend.Input{Keys: []frontend.Key{frontend.Space}})
	r.Model.Reveal = 60
	r.Model.SidebarTick = 20
	save("02-selector", r.Frame())
	r.Model.SidebarTick = 700
	save("02-selector-sidebar", r.Frame())
	for n := 1; n <= 2; n++ {
		sync(frontend.Input{Keys: []frontend.Key{frontend.Key(58 + n)}})
		r.Model.Counter = 50
		save(fmt.Sprintf("%02d-table%d-attract", 3+(n-1)*5, n), r.Frame())
		r.Model.Counter = 281
		save(fmt.Sprintf("table%d-attract-name", n), r.Frame())
		r.Model.Counter = 700
		save(fmt.Sprintf("table%d-attract-scroll", n), r.Frame())
		sync(frontend.Input{Keys: []frontend.Key{frontend.Enter}})
		for i := 0; i < 150; i++ {
			sync(frontend.Input{})
		}
		save(fmt.Sprintf("%02d-table%d-initial", 4+(n-1)*5, n), r.Frame())
		for i := 0; i < 32; i++ {
			sync(frontend.Input{Down: true})
		}
		save(fmt.Sprintf("table%d-plunger-charged", n), r.Frame())
		sync(frontend.Input{Release: true})
		save(fmt.Sprintf("table%d-plunger-released", n), r.Frame())
		for i := 0; i < 45; i++ {
			sync(frontend.Input{})
		}
		for _, held := range []bool{true, false, true, false, true, false} {
			sync(frontend.Input{Tilt: held})
		}
		for i := 0; i < 12; i++ {
			sync(frontend.Input{})
		}
		save(fmt.Sprintf("%02d-table%d-tilt", 7+(n-1)*5, n), r.Frame())
		if g, ok := r.Model.Session.(*partyland.Game); ok {
			manifest["party-state"] = map[string]any{"tilted": g.Physics.Tilted, "score": g.Score.String(), "ball": g.BallNumber}
			g.Score = partyland.Number(1234567890)
			g.Phase = partyland.GameOver
		}
		if g, ok := r.Model.Session.(*speeddevils.Game); ok {
			manifest["speed-state"] = map[string]any{"tilted": g.Physics.Tilted, "score": g.Score.String(), "ball": g.BallNumber}
			g.Score = partyland.Number(1234567890)
			g.Phase = speeddevils.GameOver
		}
		sync(frontend.Input{})
		save("table"+fmt.Sprint(n)+"-game-over", r.Frame())
		for i := 0; i < 6; i++ {
			sync(frontend.Input{})
		}
		save("table"+fmt.Sprint(n)+"-high-score", r.Frame())
		r.Model.Mode = frontend.TableAttract
		r.Model.Counter = 45
		save("table"+fmt.Sprint(n)+"-game-over-program", r.Frame())
		// No score files are written: return the in-memory model directly to selector.
		r.Model.Mode = frontend.Selector
	}
	// A captured SCROLLE half-visit at source offset 675. Glyph artwork only;
	// this checkpoint does not invent a new scheduler or host-clock mapping.
	pb, e := os.ReadFile("TABLE1.PRG")
	must(e)
	pt, e := physics.DecodePartyLand(pb)
	must(e)
	pg := partyland.New(pt, pb)
	pg.Display.Clear()
	pg.Display.GlyphText(string(pg.Display.Content.Texts["SCROLL_TEXT1"]), -675, 0, pg.Display.Content.ScrollFont)
	pp := presentation.MatrixPalette(pg.Palette(), 242)
	save("05-party-source-scroller", presentation.Compose(pg.Physics.FramePalette(pp), pg.Display, pp, 96, 242))
	// Explicit animation checkpoint consumes the existing table scheduler.
	b, e := os.ReadFile("TABLE2.PRG")
	must(e)
	t, e := physics.DecodeSpeedDevils(b)
	must(e)
	g := speeddevils.New(t, b)
	g.Display.Begin("_ANIMATION", []string{"_GEAR"})
	g.Display.Visit(4, 0, func(string) string { return "" })
	p := presentation.MatrixPalette(g.Palette(), 128)
	save("10-speed-matrix-bitmap", presentation.Compose(g.Physics.FramePalette(p), g.Display, p, 98, 128))
	raw, e := json.MarshalIndent(manifest, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(dir, "manifest.json"), append(raw, '\n'), 0644))
	fmt.Println("Native logical checkpoints:", dir)
}
func must(e error) {
	if e != nil {
		panic(e)
	}
}
