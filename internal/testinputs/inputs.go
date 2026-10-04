// Package testinputs marks tests whose external reference inputs are unavailable.
package testinputs

import (
	"os"
	"os/exec"
	"path/filepath"
	"pinballfantasies/internal/oracle"
	"testing"
)

// Generated creates private expected data from originals for this test only.
// No commercial fixture is needed in a public checkout. Missing originals or
// historical reference inputs skip; generator errors remain test failures.
func Generated(t *testing.T, script string, inputs ...string) string {
	t.Helper()
	for _, input := range inputs {
		Require(t, filepath.Join("../..", input))
	}
	output := filepath.Join(t.TempDir(), "reference.json")
	cmd := exec.Command("python3", filepath.Join("tools", script))
	cmd.Dir = "../.."
	cmd.Env = append(os.Environ(), "PF_REFERENCE_OUTPUT="+output, "PF_REFERENCE_NO_NATIVE=1")
	if log, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("private reference generation %s: %v\n%s", script, err, log)
	}
	return output
}

// Require skips only missing inputs. Unreadable or corrupt inputs still fail.
func Require(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Skipf("external original/reference input %s is absent; supply legally obtained originals or the documented private validation capture to run this integration check", path)
		} else if err != nil {
			t.Fatal(err)
		}
		// Tests using external originals are deterministic oracle checks.
		// Compatibility mutation tests deliberately read inputs independently.
		name := filepath.Base(path)
		if oracle.Known(name) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := oracle.Verify(name, data); err != nil {
				t.Fatal(err)
			}
		}
	}
}
