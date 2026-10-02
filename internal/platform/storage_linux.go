//go:build linux

package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

func InitDiagnostics(configDir string) error {
	if GUIMode != "1" {
		return nil
	}
	dir, err := StateDirectory(configDir)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "native.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	os.Stdout, os.Stderr = f, f
	return nil
}

var GUIMode string

func ReportFatal(err error) { fmt.Fprintln(os.Stderr, err) }
func RecoverFatal()         {} // Developer panics keep Go's normal stack trace.

func AudioStartupError(error) error { return nil }
