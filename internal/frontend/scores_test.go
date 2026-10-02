package frontend

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"pinballfantasies/internal/tablelogic"
)

func TestDefaultScoresSourceFidelity(t *testing.T) {
	raw, err := os.ReadFile("../../analysis/game-inventory.json")
	if err != nil {
		t.Fatal(err)
	}
	var records []struct {
		Name   string
		SHA256 string
	}
	if err := json.Unmarshal(raw, &records); err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, record := range records {
		hashes[record.Name] = record.SHA256
	}
	for table := 1; table <= 4; table++ {
		name := fmt.Sprintf("TABLE%d.HI", table)
		t.Run(name, func(t *testing.T) {
			expected := hashes[name]
			got := fmt.Sprintf("%x", sha256.Sum256(Defaults(table).MarshalBinary()))
			if expected == "" || got != expected {
				t.Fatalf("factory seed SHA256 = %s, pristine inventory = %s", got, expected)
			}
		})
	}
}

func TestFactoryScoreValues(t *testing.T) {
	names := [4][4]string{{"TSP", "ICE", "ANY", "J L"}, {"TSP", "J L", "ICE", "ANY"}, {"TSP", "ANY", "J L", "ICE"}, {"TSP", "ICE", "ANY", "J L"}}
	for table := 1; table <= 4; table++ {
		values := [4]uint64{100_000_000, 50_000_000, 25_000_000, 10_000_000}
		if table == 1 {
			values = [4]uint64{50_000_000, 25_000_000, 10_000_000, 5_000_000}
		}
		for rank, score := range Defaults(table) {
			if score.Digits.Uint64() != values[rank] || string(score.Name[:]) != names[table-1][rank] {
				t.Fatalf("table %d rank %d: %+v", table, rank, score)
			}
		}
	}
}

func TestFactoryScoresPersistenceAndReset(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprintf("TABLE%d.HI", table), func(t *testing.T) {
			root, seed := t.TempDir(), t.TempDir()
			userdata := filepath.Join(root, "userdata")
			store := FileStore{Directory: userdata, SeedDirectory: seed}
			got, err := store.Load(table)
			if err != nil || got != Defaults(table) {
				t.Fatal("fresh factory scores", got, err)
			}
			if _, err := os.Stat(userdata); !os.IsNotExist(err) {
				t.Fatal("load created mutable state", err)
			}
			got.Insert(0, tablelogic.Number(999_000_000), [3]byte{'N', 'E', 'W'})
			if err := store.Save(table, got); err != nil {
				t.Fatal(err)
			}
			reloaded, err := store.Load(table)
			if err != nil || reloaded != got {
				t.Fatal("saved record lost", err)
			}
			entries, err := os.ReadDir(userdata)
			if err != nil || len(entries) != 1 || entries[0].Name() != fmt.Sprintf("TABLE%d.HI", table) {
				t.Fatal("userdata contents", entries, err)
			}
			if entries, err := os.ReadDir(seed); err != nil || len(entries) != 0 {
				t.Fatal("seed directory changed", entries, err)
			}
			if err := os.RemoveAll(userdata); err != nil {
				t.Fatal(err)
			}
			reloaded, err = store.Load(table)
			if err != nil || reloaded != Defaults(table) {
				t.Fatal("reset factory scores", err)
			}
			// Asset-free users may still supply an optional original-format seed.
			if err := os.WriteFile(filepath.Join(seed, fmt.Sprintf("TABLE%d.HI", table)), got.MarshalBinary(), 0600); err != nil {
				t.Fatal(err)
			}
			reloaded, err = store.Load(table)
			if err != nil || reloaded != got {
				t.Fatal("optional seed ignored", err)
			}
			updated := got
			updated.Insert(0, tablelogic.Number(999_999_999), [3]byte{'U', 'S', 'R'})
			if err := store.Save(table, updated); err != nil {
				t.Fatal(err)
			}
			seedBytes, err := os.ReadFile(filepath.Join(seed, fmt.Sprintf("TABLE%d.HI", table)))
			if err != nil || string(seedBytes) != string(got.MarshalBinary()) {
				t.Fatal("optional installation seed modified", err)
			}
		})
	}
}

func TestScoresWithoutSeedDirectoryIgnoreWorkingDirectory(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	changed := Defaults(1)
	changed.Insert(0, tablelogic.Number(999_000_000), [3]byte{'L', 'O', 'C'})
	if err := os.WriteFile("TABLE1.HI", changed.MarshalBinary(), 0600); err != nil {
		t.Fatal(err)
	}
	store := FileStore{Directory: filepath.Join(t.TempDir(), "userdata")}
	got, err := store.Load(1)
	if err != nil || got != Defaults(1) {
		t.Fatal("working-directory scores used as implicit seed", got, err)
	}
}
