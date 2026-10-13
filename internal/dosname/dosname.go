// Package dosname finds original DOS file names regardless of their case.
// DOS names are case-insensitive, but copies from CDs, archives or other
// systems can arrive in lower or mixed case on case-sensitive file systems.
package dosname

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Path returns the path of name inside dir. An exact match wins. Otherwise a
// single entry whose name differs only in case is used; several such entries
// are ambiguous. Without any match the exact path is returned, so callers keep
// their ordinary not-exist error with the original name.
func Path(dir, name string) (string, error) {
	exact := filepath.Join(dir, name)
	if _, err := os.Lstat(exact); err == nil {
		return exact, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return exact, nil
	}
	found := ""
	for _, e := range entries {
		if !strings.EqualFold(e.Name(), name) {
			continue
		}
		if found != "" {
			return "", fmt.Errorf("%s: several files differ only in letter case (%s, %s)", name, found, e.Name())
		}
		found = e.Name()
	}
	if found == "" {
		return exact, nil
	}
	return filepath.Join(dir, found), nil
}

// ReadFile reads name from dir using Path.
func ReadFile(dir, name string) ([]byte, error) {
	path, err := Path(dir, name)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// Stat describes name inside dir using Path.
func Stat(dir, name string) (os.FileInfo, error) {
	path, err := Path(dir, name)
	if err != nil {
		return nil, err
	}
	return os.Stat(path)
}
