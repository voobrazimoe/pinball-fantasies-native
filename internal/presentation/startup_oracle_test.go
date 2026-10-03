package presentation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// Export only hashes and source metadata. The owner-supplied DOS screenshots
// and linked content remain private. The first clear may retain part of the
// preceding scroll; feed its sampled predecessor dot grids through the source
// selection program rather than assuming a blank startup framebuffer.
func TestExportStartupOracleStates(t *testing.T) {
	path := os.Getenv("PF_MATRIX_STARTUP_STATES")
	if path == "" {
		t.Skip("private one-player DOS capture comparison")
	}
	table, err := strconv.Atoi(os.Getenv("PF_MATRIX_ORACLE_TABLE"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(os.Getenv("PF_MATRIX_PREDECESSOR_GRIDS"))
	if err != nil {
		t.Fatal(err)
	}
	var grids [][]byte
	if err = json.Unmarshal(raw, &grids); err != nil {
		t.Fatal(err)
	}
	all := map[string][]map[string]interface{}{}
	for predecessor, grid := range grids {
		d := original(t, table)
		if len(grid) != len(d.Dots) {
			t.Fatal("invalid predecessor dots")
		}
		for i, b := range grid {
			d.Dots[i] = b != 0
		}
		d.SetPlayers(1, 1, 1)
		r := replay{d: d, commands: d.Content.Commands, pc: d.Content.Labels["FIRST_NO_OF_PLAYERSTS"], active: true}
		d.StartMatrix()
		r.dispatch()
		for tick := 0; tick < 150; tick++ {
			dots := make([]byte, len(d.Dots))
			for i, b := range d.Dots {
				if b {
					dots[i] = 1
				}
			}
			hash := fmt.Sprintf("%x", sha256.Sum256(dots))
			all[hash] = append(all[hash], map[string]interface{}{"tick": tick, "initial_phase": predecessor, "op": d.op})
			if d.op == "_WAIT_GAME_ON" {
				d.FlushPrint(r.number)
			} else {
				r.step()
			}
		}
	}
	b, err := json.Marshal(all)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
}
