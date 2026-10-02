package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func DefaultDataDir() string { return "" }

// Resolve the outer executable, not CWD or an AppImage's temporary mount.
func portableRoot(executable, appImage string, linux bool) (string, error) {
	if linux && appImage != "" {
		executable = appImage
	}
	p, err := filepath.Abs(executable)
	if err != nil {
		return "", err
	}
	p, err = filepath.EvalSymlinks(p)
	if err != nil {
		return "", fmt.Errorf("portable executable path: %w", err)
	}
	return filepath.Dir(p), nil
}

func PortableRoot() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return portableRoot(exe, os.Getenv("APPIMAGE"), runtime.GOOS == "linux")
}

func ResolveDataDir(value string) (string, error) {
	if value != "" {
		return filepath.Abs(value)
	}
	return PortableRoot()
}

func writableDirectory(dir string) (string, error) {
	p, err := filepath.Abs(dir)
	if err == nil {
		err = os.MkdirAll(p, 0700)
	}
	if err == nil {
		var f *os.File
		f, err = os.CreateTemp(p, ".native-write-check-*")
		if err == nil {
			_, err = f.Write([]byte("portable storage check\n"))
			ce := f.Close()
			re := os.Remove(f.Name())
			if err == nil {
				err = ce
			}
			if err == nil {
				err = re
			}
		}
	}
	if err != nil {
		return "", fmt.Errorf("cannot write native state in %q: %w; move the game to a writable directory or use -config-dir and -high-score-dir", dir, err)
	}
	return filepath.EvalSymlinks(p)
}

func UserStorageDir() (string, error) {
	root, err := PortableRoot()
	if err != nil {
		return "", err
	}
	return portableStorage(root, os.UserConfigDir)
}

func portableStorage(root string, userConfig func() (string, error)) (string, error) {
	dir, portableErr := writableDirectory(filepath.Join(root, "userdata"))
	if portableErr == nil {
		return dir, nil
	}
	base, err := userConfig()
	if err != nil {
		return "", fmt.Errorf("%v; user configuration fallback: %w", portableErr, err)
	}
	name := "pinballfantasies"
	if runtime.GOOS == "windows" {
		name = "PinballFantasies"
	}
	dir, err = writableDirectory(filepath.Join(base, name))
	if err != nil {
		return "", fmt.Errorf("%v; user configuration fallback: %w", portableErr, err)
	}
	return dir, nil
}

// Explicit overrides are checked too, before gameplay can mutate state.
func StateDirectory(override string) (string, error) {
	if override != "" {
		return writableDirectory(override)
	}
	return UserStorageDir()
}

// Supplied legacy CFG/HI names also name native writable files. Reject aliases
// of the data root so an explicit override cannot overwrite the originals.
func CheckStateSeparate(dataDir, stateDir string) error {
	data, err := filepath.EvalSymlinks(dataDir)
	if err != nil {
		return err
	}
	state, err := filepath.EvalSymlinks(stateDir)
	if err != nil {
		return err
	}
	same := data == state
	if runtime.GOOS == "windows" {
		same = strings.EqualFold(data, state)
	}
	if same {
		return fmt.Errorf("native state directory must differ from original data directory %q; use userdata or another -config-dir/-high-score-dir", dataDir)
	}
	return nil
}
