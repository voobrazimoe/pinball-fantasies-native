package frontend

import (
	"fmt"
	"os"
	"path/filepath"
	"pinballfantasies/internal/datalayout"
	"pinballfantasies/internal/tablelogic"
	"testing"
)

func TestPrivateInstallationFactoryInitialsPersistence(t *testing.T) {
	for _, env := range []string{"PF_RUNTIME_DATA", "PF_POWERPACK_DATA", "PF_DELUXE_CD_DATA", "PF_DELUXE_CD_ALT_DATA"} {
		t.Run(env, func(t *testing.T) {
			dir := os.Getenv(env)
			if dir == "" {
				t.Skip("supply " + env)
			}
			stage := stageInstallation(t, dir)
			state := t.TempDir()
			store := FileStore{Directory: state, SeedDirectory: t.TempDir()}
			r, e := Load(stage, store)
			if e != nil {
				t.Fatal(e)
			}
			nilStore, e := Load(stage, nil)
			if e != nil {
				t.Fatal(e)
			}
			var expected [4]Scores
			for table := 1; table <= 4; table++ {
				name := fmt.Sprintf("TABLE%d.PRG", table)
				src, e := os.ReadFile(filepath.Join(dir, name))
				if e != nil {
					t.Fatal(e)
				}
				decoded, e := datalayout.PreparePRGForProfile(r.ProfileID, name, src)
				if e != nil {
					t.Fatal(e)
				}
				names, e := datalayout.FactoryInitials(name, decoded)
				if e != nil {
					t.Fatal(e)
				}
				expected[table-1] = Defaults(table)
				for rank := range names {
					expected[table-1][rank].Name = names[rank]
				}
				if env == "PF_DELUXE_CD_ALT_DATA" {
					for rank := range names {
						want := []string{"CLS", "RLZ", "CLS", "RLZ"}[rank]
						if string(names[rank][:]) != want {
							t.Fatal("C canonicalized", table, names)
						}
					}
				} else if expected[table-1] != Defaults(table) {
					t.Fatal("A/B/D changed")
				}
				if r.Model.Scores[table-1] != expected[table-1] || nilStore.Model.Scores[table-1] != expected[table-1] {
					t.Fatal("initial seed lost", table)
				}
				saved := expected[table-1]
				saved.Insert(0, tablelogic.Number(999000000), [3]byte{'N', 'E', 'W'})
				if e := r.Model.Store.Save(table, saved); e != nil {
					t.Fatal(e)
				}
				restarted, e := Load(stage, store)
				if e != nil || restarted.Model.Scores[table-1] != saved {
					t.Fatal("restart", e)
				}
				if e := os.Remove(filepath.Join(state, fmt.Sprintf("TABLE%d.HI", table))); e != nil {
					t.Fatal(e)
				}
				reset, e := Load(stage, store)
				if e != nil || reset.Model.Scores[table-1] != expected[table-1] {
					t.Fatal("clear/reset", e)
				}
			}
			// Explicit optional seeds remain user state, including across editions.
			if e := os.WriteFile(filepath.Join(store.SeedDirectory, "TABLE2.HI"), Defaults(2).MarshalBinary(), 0600); e != nil {
				t.Fatal(e)
			}
			seeded, e := Load(stage, store)
			if e != nil || seeded.Model.Scores[1] != Defaults(2) {
				t.Fatal("optional seed precedence", e)
			}
			if e := os.Remove(filepath.Join(store.SeedDirectory, "TABLE2.HI")); e != nil {
				t.Fatal(e)
			}
			// One common state namespace is intentional. Existing saved scores win
			// across installation changes, while missing records use the new defaults.
			if other := os.Getenv("PF_DELUXE_CD_DATA"); env == "PF_DELUXE_CD_ALT_DATA" && other != "" {
				if e := r.Model.Store.Save(1, expected[0]); e != nil {
					t.Fatal(e)
				}
				d, e := Load(stageInstallation(t, other), store)
				if e != nil || d.Model.Scores[0] != expected[0] {
					t.Fatal("saved state precedence", e)
				}
				if d.Model.Scores[1] != Defaults(2) {
					t.Fatal("global C default contamination")
				}
			}
		})
	}
}
