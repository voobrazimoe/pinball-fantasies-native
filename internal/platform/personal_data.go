package platform

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PrepareDataDir keeps bundled originals separate from PortableRoot and native
// writable state. An explicit directory overrides the personal bundle.
func PrepareDataDir(value string) (string, func(), error) {
	if value == "" && len(personalPayload) != 0 {
		return extractPersonalData(personalPayload)
	}
	dir, err := ResolveDataDir(value)
	return dir, func() {}, err
}

func extractPersonalData(payload []byte) (string, func(), error) {
	archive, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return "", nil, fmt.Errorf("personal data bundle: %w", err)
	}
	dir, err := os.MkdirTemp("", "pinballfantasies-private-data-*")
	if err != nil {
		return "", nil, fmt.Errorf("personal data temporary directory: %w", err)
	}
	cleanup := func() {
		// Windows read-only attributes must be cleared for clean-exit deletion.
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			os.Chmod(filepath.Join(dir, entry.Name()), 0600)
		}
		if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintln(os.Stderr, "personal data cleanup:", err)
		}
	}
	fail := func(err error) (string, func(), error) { cleanup(); return "", nil, err }
	for _, entry := range archive.File {
		// Build bundles contain flat, allowlisted filenames only. Reject traversal
		// and nested names independently at runtime before using an archive path.
		if entry.Name == "" || entry.Name == "." || entry.Name == ".." || strings.ContainsAny(entry.Name, "/\\:") || entry.UncompressedSize64 > 16<<20 {
			return fail(fmt.Errorf("invalid personal data entry %q", entry.Name))
		}
		r, err := entry.Open()
		if err != nil {
			return fail(err)
		}
		path := filepath.Join(dir, entry.Name)
		w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			r.Close()
			return fail(err)
		}
		_, err = io.Copy(w, r)
		r.Close()
		ce := w.Close()
		if err == nil {
			err = ce
		}
		if err == nil {
			err = os.Chmod(path, 0400)
		}
		if err != nil {
			return fail(err)
		}
	}
	return dir, cleanup, nil
}
