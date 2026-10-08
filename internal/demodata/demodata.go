// Package demodata carries the official 10-minute Party Land demo, which public
// builds ship so a player without the full game can still try it.
package demodata

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed assets/demo/*
var files embed.FS

// Files are the bundled demo files with their SHA-256 identities. PINBALL.EXE
// only supplies the closing text; it is never executed.
var Files = []struct{ Name, SHA256 string }{
	{"INTRO.PRG", "05bdba35e0a9a31a87b944428ddad27a00ba8983e97963290ab2ee1f57910fa3"},
	{"INTRO.MOD", "f36beae00efec1dd9e1c4e977bea264b7ec41ab18f258528ab66577a9ec66613"},
	{"MOD2.MOD", "aa5003c275b494062f37f44e8c77105b8a420555f4bd6ff53d7698f89c540f21"},
	{"TABLE1.PRG", "44b8f4b76ee16c47cda26904681e83e8f4420974bcea249d099790bfbd7369e3"},
	{"TABLE1.MOD", "a0877e4372abe64b70d9e361bf257ea5a84c948771f0eace3433d5f6399060b5"},
	{"PINBALL.EXE", "7acc8be42f23cc56a66ce8839a561c4d7318025dc395935be44b7a760f538dff"},
}

// Extract writes the demo into a new temporary directory. The cleanup removes it.
func Extract() (string, func(), error) {
	dir, err := os.MkdirTemp("", "pinballfantasies-demo-*")
	if err != nil {
		return "", nil, fmt.Errorf("demo temporary directory: %w", err)
	}
	cleanup := func() {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintln(os.Stderr, "demo cleanup:", err)
		}
	}
	for _, f := range Files {
		b, err := files.ReadFile("assets/demo/" + f.Name)
		if err == nil {
			sum := sha256.Sum256(b)
			if hex.EncodeToString(sum[:]) != f.SHA256 {
				err = fmt.Errorf("bundled demo file %s is damaged", f.Name)
			}
		}
		if err == nil {
			err = os.WriteFile(filepath.Join(dir, f.Name), b, 0600)
		}
		if err != nil {
			cleanup()
			return "", nil, err
		}
	}
	return dir, cleanup, nil
}
