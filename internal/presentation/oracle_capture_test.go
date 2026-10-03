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

func TestExportScrollOracleStates(t *testing.T) {
	path := os.Getenv("PF_MATRIX_ORACLE_STATES")
	if path == "" {
		t.Skip("private dynamic DOS capture export")
	}
	table := 1
	if v := os.Getenv("PF_MATRIX_ORACLE_TABLE"); v != "" {
		var err error
		table, err = strconv.Atoi(v)
		if err != nil || table < 1 || table > 4 {
			t.Fatal("PF_MATRIX_ORACLE_TABLE must be 1..4")
		}
	}
	d := original(t, table)
	root := os.Getenv("PF_MATRIX_ORACLE_PAYLOAD")
	if root == "" {
		root = "../../.cache/matrix-original-oracle"
	}
	hi, e := os.ReadFile(filepath.Join(root, fmt.Sprintf("TABLE%d.HI", table)))
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range d.Content.Attract {
		if strings.HasPrefix(c.Op, "_PRINT") && strings.Contains(c.Arg(0), "HI_SCORE_LIST") {
			for rank := 0; rank < 4; rank++ {
				if strings.Contains(c.Arg(0), fmt.Sprintf("*%d+12", rank)) {
					d.Content.Texts[c.Arg(0)] = append(append([]byte{}, hi[rank*16+12:rank*16+15]...), 0)
				}
			}
		}
	}
	number := func(label string) string {
		for rank := 0; rank < 4; rank++ {
			if strings.Contains(label, fmt.Sprintf("*%d)", rank)) {
				var s strings.Builder
				for _, v := range hi[rank*16 : rank*16+12] {
					s.WriteByte('0' + v)
				}
				return s.String()
			}
		}
		return "0"
	}
	all := map[string][]map[string]interface{}{}
	grids := map[string][]byte{}
	for initial := uint8(1); initial <= 8; initial++ {
		out := *d
		out.scrollPhase = initial
		out.scrollPlane = [2]int{int(initial % 2), 1 - int(initial%2)}
		r := replay{d: &out, commands: append(append([]Command{}, out.Content.Attract...), Command{Op: "0"}), active: true, numberFunc: number}
		r.dispatch()
		for tick := 0; tick < 8000; tick++ {
			dots := make([]byte, 2560)
			for i, v := range out.Dots {
				if v {
					dots[i] = 1
				}
			}
			hash := fmt.Sprintf("%x", sha256.Sum256(dots))
			grids[hash] = dots
			all[hash] = append(all[hash], map[string]interface{}{"tick": tick, "initial_phase": initial, "op": out.op, "phase": out.scrollPhase, "char": out.scrollChar})
			r.step()
			if !r.active {
				r = replay{d: &out, commands: append(append([]Command{}, out.Content.Attract...), Command{Op: "0"}), active: true, numberFunc: number}
				r.dispatch()
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	gridJSON, err := json.Marshal(grids)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path+".grids", gridJSON, 0644); err != nil {
		t.Fatal(err)
	}
	b, e := json.Marshal(all)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, b, 0644); e != nil {
		t.Fatal(e)
	}
}
