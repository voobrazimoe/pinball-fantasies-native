//go:build windows

package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"unsafe"
)

// GUIMode is set only by the release build wrapper. Debug builds retain stderr/panics.
var GUIMode string
var diagnosticFile *os.File

func InitDiagnostics(configDir string) error {
	if GUIMode != "1" {
		return nil
	}
	dir, e := StateDirectory(configDir)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(dir, 0700); e != nil {
		return e
	}
	diagnosticFile, e = os.OpenFile(filepath.Join(dir, "native.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if e != nil {
		return e
	}
	// Set native standard handles too: Go runtime panic/fatal diagnostics write
	// through GetStdHandle, rather than the os.Stderr variable.
	for _, id := range []uintptr{^uintptr(10), ^uintptr(11)} {
		if v, _, e := kernel32.NewProc("SetStdHandle").Call(id, diagnosticFile.Fd()); v == 0 {
			return apiError("diagnostic SetStdHandle", e)
		}
	}
	os.Stdout = diagnosticFile
	os.Stderr = diagnosticFile
	return nil
}
func ReportFatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	if GUIMode == "1" {
		up("MessageBoxW").Call(0, uintptr(unsafe.Pointer(utf(err.Error()+"\n\nOriginal data: use -data-dir <installation directory>. Diagnostics: userdata\\native.log beside the executable (or -config-dir)."))), uintptr(unsafe.Pointer(utf("Pinball Fantasies"))), 0x10)
	}
}
func RecoverFatal() {
	if GUIMode != "1" {
		return
	}
	if p := recover(); p != nil {
		fmt.Fprintf(os.Stderr, "panic: %v\n%s", p, debug.Stack())
		ReportFatal(fmt.Errorf("unexpected failure: %v", p))
		os.Exit(1)
	}
}

func AudioStartupError(e error) error { return e }
