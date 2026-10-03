package presentation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Opt-in private DOS ending comparison. Export only hashes/source metadata;
// supplied retail high scores and screenshots never enter the public tree.
func TestExportEndingOracleStates(t *testing.T) {
	path := os.Getenv("PF_MATRIX_ENDING_STATES")
	if path == "" {
		t.Skip("private DOS ending oracle")
	}
	table, err := strconv.Atoi(os.Getenv("PF_MATRIX_ORACLE_TABLE"))
	if err != nil || table < 1 || table > 4 {
		t.Fatal("table must be 1..4")
	}
	hi, err := os.ReadFile(filepath.Join(os.Getenv("PF_MATRIX_ORACLE_PAYLOAD"), fmt.Sprintf("TABLE%d.HI", table)))
	if err != nil {
		t.Fatal(err)
	}
	var names, scores [4]string
	for rank := range names {
		names[rank] = string(hi[rank*16+12 : rank*16+15])
		for _, v := range hi[rank*16 : rank*16+12] {
			scores[rank] += string('0' + v)
		}
	}
	all := map[string][]map[string]interface{}{}
	for phase := uint8(1); phase <= 8; phase++ {
		d := original(t, table)
		d.scrollPhase = phase
		d.scrollPlane = [2]int{int(phase % 2), 1 - int(phase%2)}
		d.BeginResolved("_MATRIXLGT", []string{"0"}, map[int]int{0: 0})
		d.Visit(0, 0, func(string) string { return "0" })
		timeline := d.ContinueGameOver([]string{strings.Repeat("0", 12)}, names, scores)
		for tick := 0; tick < 5000; tick++ {
			timeline.Tick()
			dots := make([]byte, 2560)
			for i, v := range timeline.Display().Dots {
				if v {
					dots[i] = 1
				}
			}
			hash := fmt.Sprintf("%x", sha256.Sum256(dots))
			all[hash] = append(all[hash], map[string]interface{}{"tick": tick, "initial_phase": phase, "op": timeline.Display().op})
		}
	}
	b, err := json.Marshal(all)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
}
