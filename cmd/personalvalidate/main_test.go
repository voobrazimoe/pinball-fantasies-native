package main

import (
	"os"
	"path/filepath"
	"pinballfantasies/internal/testinputs"
	"testing"
)

func TestValidationIgnoresAmbientHighScores(t *testing.T) {
	data := t.TempDir()
	names := []string{"INTRO.PRG", "INTRO.MOD", "MOD2.MOD", "TABLE1.PRG", "TABLE1.MOD", "TABLE2.PRG", "TABLE2.MOD", "TABLE3.PRG", "TABLE3.MOD", "TABLE4.PRG", "TABLE4.MOD"}
	for _, name := range names {
		testinputs.Require(t, filepath.Join("../..", name))
		raw, err := os.ReadFile(filepath.Join("../..", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(data, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	cwd := t.TempDir()
	t.Chdir(cwd)
	for _, dir := range []string{data, cwd} {
		for _, name := range []string{"TABLE1.HI", "TABLE2.HI", "TABLE3.HI", "TABLE4.HI"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("invalid ambient scores"), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := validate(data); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{data, cwd} {
		for _, name := range []string{"TABLE1.HI", "TABLE2.HI", "TABLE3.HI", "TABLE4.HI"} {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil || string(raw) != "invalid ambient scores" {
				t.Fatal("ambient seed modified", err)
			}
			if err := os.Remove(filepath.Join(dir, name)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := validate(data); err != nil {
		t.Fatal(err)
	}
}
