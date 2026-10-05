// pflayoutaudit validates every possessed input independently; never emits payload.
package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/settings"
)

type result struct {
	Name     string
	Accepted bool
	Reason   string `json:",omitempty"`
}

func main() {
	dir := flag.String("data", "", "private DOS input directory (read only)")
	flag.Parse()
	if *dir == "" {
		flag.Usage()
		os.Exit(2)
	}
	checks := []struct {
		name  string
		check func([]byte) error
	}{
		{"INTRO.PRG", func(b []byte) error {
			c, e := datalayout.PreparePRG("INTRO.PRG", b)
			if e != nil {
				return e
			}
			_, e = assets.DecodeFrontend(c)
			return e
		}},
	}
	prgs := []func([]byte) error{
		func(b []byte) error { _, e := physics.DecodePartyLand(b); return e },
		func(b []byte) error { _, e := physics.DecodeSpeedDevils(b); return e },
		func(b []byte) error { _, e := physics.DecodeGameshow(b); return e },
		func(b []byte) error { _, e := physics.DecodeStones(b); return e },
	}
	for i, n := range []string{"TABLE1.PRG", "TABLE2.PRG", "TABLE3.PRG", "TABLE4.PRG"} {
		decode := prgs[i]
		name := n
		checks = append(checks, struct {
			name  string
			check func([]byte) error
		}{name, func(b []byte) error {
			c, e := datalayout.PreparePRG(name, b)
			if e != nil {
				return e
			}
			return decode(c)
		}})
	}
	for i, n := range []string{"INTRO.MOD", "MOD2.MOD", "TABLE1.MOD", "TABLE2.MOD", "TABLE3.MOD", "TABLE4.MOD"} {
		decode := []func([]byte) (*audio.Module, error){audio.DecodeIntro, audio.DecodeMenu, audio.Decode, audio.DecodeSpeedDevils, audio.DecodeGameshow, audio.DecodeStones}[i]
		checks = append(checks, struct {
			name  string
			check func([]byte) error
		}{n, func(b []byte) error { _, e := decode(b); return e }})
	}
	checks = append(checks, struct {
		name  string
		check func([]byte) error
	}{"PINBALL.CFG", func(b []byte) error { _, e := settings.Decode(b); return e }})
	out := []result{}
	for _, c := range checks {
		b, e := os.ReadFile(filepath.Join(*dir, c.name))
		if c.name == "PINBALL.CFG" && os.IsNotExist(e) {
			out = append(out, result{c.name, true, "optional; absent"})
			continue
		}
		if e == nil {
			e = c.check(b)
		}
		r := result{Name: c.name, Accepted: e == nil}
		if e != nil {
			r.Reason = e.Error()
		}
		out = append(out, r)
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
