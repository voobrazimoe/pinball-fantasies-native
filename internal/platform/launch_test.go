//go:build linux || windows

package platform

import (
	"os"
	"path/filepath"
	"testing"

	"pinballfantasies/internal/demodata"
)

func scripted(choices ...launchChoice) func() launchChoice {
	return func() launchChoice { c := choices[0]; choices = choices[1:]; return c }
}

func TestLaunchDemoLeavesImportEmpty(t *testing.T) {
	imported := filepath.Join(t.TempDir(), "Data")
	dir, cleanup, err := chooseLaunch(imported, scripted(launchDemo), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if _, err := os.Stat(filepath.Join(dir, "TABLE1.PRG")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(imported); !os.IsNotExist(err) {
		t.Fatal("demo must not count as an imported game")
	}
}

func TestLaunchRejectsDemoAsFullGame(t *testing.T) {
	demo, cleanup, err := demodata.Extract()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	imported := filepath.Join(t.TempDir(), "Data")
	var shown string
	_, _, err = chooseLaunch(imported, scripted(launchImport, launchQuit),
		func() (string, bool) { return demo, true }, func(s string) { shown = s })
	if err != ErrQuit || shown == "" {
		t.Fatalf("err %v, message %q", err, shown)
	}
	if _, err := os.Stat(imported); !os.IsNotExist(err) {
		t.Fatal("rejected import left files behind")
	}
}

func TestLaunchImportsFullGame(t *testing.T) {
	source := os.Getenv("PF_POWERPACK_DATA")
	if source == "" {
		source = os.Getenv("PF_RUNTIME_DATA")
	}
	if source == "" {
		t.Skip("supply a private full installation")
	}
	imported := filepath.Join(t.TempDir(), "Data")
	for range 2 { // the second import replaces the first
		dir, _, err := chooseLaunch(imported, scripted(launchImport),
			func() (string, bool) { return source, true }, func(s string) { t.Fatal(s) })
		if err != nil || dir != imported {
			t.Fatal(dir, err)
		}
		if err := fullGame(imported); err != nil {
			t.Fatal(err)
		}
	}
	entries, _ := os.ReadDir(filepath.Dir(imported))
	if len(entries) != 1 {
		t.Fatalf("staging left behind: %v", entries)
	}
}
