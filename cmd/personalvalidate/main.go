// Personal packaging validation uses the existing read-only runtime decoders.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/tablelogic"
)

func main() {
	dir := flag.String("data-dir", "", "original personal data directory")
	flag.Parse()
	if err := validate(*dir); err != nil {
		fmt.Fprintln(os.Stderr, "personal data validation:", err)
		os.Exit(1)
	}
	fmt.Println("11 game-data inputs and optional settings validated; all four factory scores save, restart and reset in isolated storage")
}

func validate(dir string) error {
	isolated, err := os.MkdirTemp("", "personal-validation-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(isolated)
	// Explicit empty seed and writable paths never consult installation or CWD.
	seed := filepath.Join(isolated, "seed")
	if err := os.Mkdir(seed, 0700); err != nil {
		return err
	}
	store := frontend.FileStore{Directory: filepath.Join(isolated, "userdata"), SeedDirectory: seed}
	runtime, err := frontend.Load(dir, store)
	if err != nil {
		return err
	}
	for table := 1; table <= 4; table++ {
		factory := frontend.Defaults(table)
		if runtime.Model.Scores[table-1] != factory {
			return fmt.Errorf("table %d: not factory defaults", table)
		}
		saved := factory
		saved.Insert(0, tablelogic.Number(999_000_000), [3]byte{'N', 'E', 'W'})
		if err := store.Save(table, saved); err != nil {
			return err
		}
		restarted := frontend.FileStore{Directory: store.Directory, SeedDirectory: seed}
		got, err := restarted.Load(table)
		if err != nil {
			return err
		}
		if got != saved {
			return fmt.Errorf("table %d: saved score lost", table)
		}
		if err := os.Remove(filepath.Join(store.Directory, fmt.Sprintf("TABLE%d.HI", table))); err != nil {
			return err
		}
		got, err = restarted.Load(table)
		if err != nil {
			return err
		}
		if got != factory {
			return fmt.Errorf("table %d: reset failed", table)
		}
	}
	entries, err := os.ReadDir(seed)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("validation modified seed directory")
	}
	return nil
}
