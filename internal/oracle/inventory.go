// Package oracle verifies exact private reference inputs. Runtime hosts must
// use the compatible data decoders, never this research/fixture identity gate.
package oracle

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed inventory.json
var inventory []byte
var pins = func() map[string]struct {
	Size   int
	SHA256 string
} { var records []struct {
	Name   string
	Size   int
	SHA256 string
}; if err := json.Unmarshal(inventory, &records); err != nil {
	panic(err)
}; out := make(map[string]struct {
	Size   int
	SHA256 string
}); for _, r := range records {
	out[r.Name] = struct {
		Size   int
		SHA256 string
	}{r.Size, r.SHA256}
}; return out }()

func Known(name string) bool { _, ok := pins[name]; return ok }
func Verify(name string, data []byte) error {
	p, ok := pins[name]
	if !ok {
		return fmt.Errorf("no pinned oracle for %s", name)
	}
	if len(data) != p.Size || fmt.Sprintf("%x", sha256.Sum256(data)) != p.SHA256 {
		return fmt.Errorf("%s differs from exact pinned oracle", name)
	}
	return nil
}
