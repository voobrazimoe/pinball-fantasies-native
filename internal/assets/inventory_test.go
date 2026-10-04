package assets

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"pinballfantasies/internal/testinputs"
	"strings"
	"testing"
)

func TestOriginalInventories(t *testing.T) {
	for _, group := range []struct{ manifest, root string }{
		{"game-inventory.json", "../.."},
		{"source-inventory.json", "../../reference/original-dos-source"},
	} {
		raw, err := os.ReadFile(filepath.Join("../../analysis", group.manifest))
		if err != nil {
			t.Fatal(err)
		}
		var records []struct {
			Name   string `json:"name"`
			Size   int    `json:"size"`
			SHA256 string `json:"sha256"`
		}
		if err := json.Unmarshal(raw, &records); err != nil {
			t.Fatal(err)
		}
		for _, r := range records {
			// Mutable DOS settings and scores are optional seeds. Score pristine hashes are
			// verified against frontend.Defaults by TestDefaultScoresSourceFidelity.
			if group.manifest == "game-inventory.json" && (r.Name == "PINBALL.CFG" || strings.HasSuffix(r.Name, ".HI")) {
				continue
			}
			t.Run(group.manifest+"/"+r.Name, func(t *testing.T) {
				testinputs.Require(t, filepath.Join(group.root, r.Name))
				b, err := os.ReadFile(filepath.Join(group.root, r.Name))
				if err != nil {
					t.Fatal(err)
				}
				if len(b) != r.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != r.SHA256 {
					t.Fatal("original file differs from pinned inventory")
				}
			})
		}
	}
}
