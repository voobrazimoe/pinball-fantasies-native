// Source-state checkpoints and continuous frontend PCM, without a DOS runtime.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/presentation"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	out := "analysis/pf8-runtime-validation/native"
	must(os.MkdirAll(out, 0755))
	r, e := frontend.Load(".", frontend.FileStore{Directory: ".", SeedDirectory: "."})
	must(e)
	var records []map[string]any
	save := func(name string) {
		f, e := os.Create(filepath.Join(out, name+".png"))
		must(e)
		im := r.Frame()
		must(png.Encode(f, im))
		must(f.Close())
		records = append(records, map[string]any{"name": name, "hash": fmt.Sprintf("%x", sha256.Sum256(im.Pix)), "mode": r.Model.Mode.String(), "tick": r.Model.Tick, "segment": r.Model.Segment, "segment_tick": r.Model.SegmentTick, "reveal": r.Model.Reveal, "text_tick": r.Model.TextTick, "text_page": r.Model.TextPage, "exit_tick": r.Model.TextExitTick, "sidebar": r.Model.SidebarTick})
	}
	// Every distinct startup fade frame; holds reuse the same logical pixels.
	lastHash := ""
	for r.Model.Mode == frontend.Startup {
		m := r.Model
		hash := fmt.Sprintf("%x", sha256.Sum256(r.Frame().Pix))
		if hash != lastHash {
			save(fmt.Sprintf("startup-%02d-%03d", m.Segment, m.SegmentTick))
			lastHash = hash
		}
		must(r.Update(frontend.Input{}))
	}
	for i := 0; i < 65; i++ {
		if i%4 == 0 || i == 17 || i == 40 {
			save(fmt.Sprintf("selector-%02d", i))
		}
		must(r.Update(frontend.Input{}))
	}
	must(r.Update(frontend.Input{Keys: []frontend.Key{frontend.Space}}))
	for i := 0; i < 46; i++ {
		save(fmt.Sprintf("text-reveal-%02d", i))
		must(r.Update(frontend.Input{}))
	}
	r.Model.Counter = 0
	for i := 0; i < 22; i++ {
		save(fmt.Sprintf("text-exit-%02d", i))
		must(r.Update(frontend.Input{}))
	}
	var matrices []map[string]any
	for n := 1; n <= 2; n++ {
		base := r.View.Matrix
		on, off := byte(242), byte(96)
		if n == 2 {
			base = r.View.SpeedMatrix
			on, off = 128, 98
		}
		var names, scores [4]string
		for i, v := range r.Model.Scores[n-1] {
			names[i] = string(v.Name[:])
			scores[i] = v.Digits.String()
		}
		var pal [768]byte
		for i := 0; i < 3; i++ {
			pal[int(off)*3+i] = assets.DACRGB(20)
		}
		pal = presentation.MatrixPalette(pal, on)
		for tick := 0; tick < 1000; tick++ {
			d := base.Attract(tick, names, scores)
			dots := make([]byte, 2560)
			for i, v := range d.Dots {
				if v {
					dots[i] = 1
				}
			}
			im := image.NewRGBA(image.Rect(0, 0, 320, 350))
			d.Draw(im, pal, off, on)
			matrices = append(matrices, map[string]any{"table": n, "tick": tick, "on": d.On, "dots": fmt.Sprintf("%x", sha256.Sum256(dots)), "rgba": fmt.Sprintf("%x", sha256.Sum256(im.Pix[317*im.Stride:]))})
		}
	}
	matrixJSON, e := json.MarshalIndent(matrices, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(out, "attract-states.json"), matrixJSON, 0644))
	raw, e := json.MarshalIndent(records, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(out, "states.json"), raw, 0644))
	p := audio.New(r.Intro)
	must(os.WriteFile("analysis/pf8-runtime-validation/selector-expected.pcm", p.Render(audio.Rate*24), 0644))
}
