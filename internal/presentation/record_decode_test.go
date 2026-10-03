package presentation

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"pinballfantasies/internal/gameplay"
	"regexp"
	"sort"
	"strconv"
	"testing"
)

func fingerprintRecords(c Content, category string) string {
	h := sha256.New()
	put := func(n int) { var b [8]byte; binary.LittleEndian.PutUint64(b[:], uint64(n)); h.Write(b[:]) }
	keys := []string{}
	switch category {
	case "texts":
		for k := range c.Texts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			h.Write([]byte(k + "\x00"))
			put(len(c.Texts[k]))
			h.Write(c.Texts[k])
		}
	case "animations":
		for k := range c.Animations {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			h.Write([]byte(k + "\x00"))
			a := c.Animations[k]
			put(len(a.Header))
			for _, v := range a.Header {
				put(int(v))
			}
			put(len(a.Durations))
			for _, v := range a.Durations {
				put(int(v))
			}
			put(len(a.Offsets))
			for _, v := range a.Offsets {
				put(v)
			}
		}
	case "lampflash":
		for _, r := range c.LampFlash {
			for _, v := range r {
				put(v)
			}
		}
	}
	return hashString(h)
}
func hashString(h hash.Hash) string { return fmt.Sprintf("%x", h.Sum(nil)) }

// These are pre-migration accepted record fingerprints, never commercial
// expected output. Independent ASM text/writer checks also run with originals.
func TestPRGRecordBaselineParity(t *testing.T) {
	var proof map[string]map[string]struct {
		Count  int
		SHA256 string
	}
	raw, err := os.ReadFile("../../analysis/prg-record-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &proof); err != nil {
		t.Fatal(err)
	}
	for n := 1; n <= 4; n++ {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			d := New(n, prg(t, n))
			// Keep the historical table-local record fingerprint intact. The
			// previously omitted FANTASIE records are tested independently.
			baseline := d.Content
			baseline.Texts = make(map[string][]byte, len(d.Content.Texts))
			for k, v := range d.Content.Texts {
				baseline.Texts[k] = v
			}
			for _, cheat := range gameplay.OriginalCheats {
				first := d.Content.Labels[cheat.Program]
				delete(baseline.Texts, d.Content.Commands[first+1].Arg(0))
			}
			for rank := 1; rank <= 4; rank++ {
				delete(baseline.Texts, "HI_"+strconv.Itoa(rank))
			}
			for category, p := range proof[strconv.Itoa(n)] {
				count := len(baseline.Texts)
				if category == "animations" {
					count = len(d.Content.Animations)
				}
				if category == "lampflash" {
					count = len(d.Content.LampFlash)
				}
				if count != p.Count || fingerprintRecords(baseline, category) != p.SHA256 {
					t.Fatalf("table %d %s baseline parity", n, category)
				}
			}
		})
	}
}

func TestSyntheticRecordDecode(t *testing.T) {
	b := make([]byte, 60)
	copy(b[2:], []byte("NEW\x00"))
	binary.LittleEndian.PutUint16(b[10:], 3)
	binary.LittleEndian.PutUint16(b[12:], 2)
	binary.LittleEndian.PutUint16(b[14:], 7)
	binary.LittleEndian.PutUint16(b[16:], 30)
	binary.LittleEndian.PutUint16(b[18:], 9)
	for j, v := range []uint16{5, 6, 7, 8} {
		binary.LittleEndian.PutUint16(b[42+j*2:], v)
	}
	c := Content{Texts: map[string][]byte{}, TextRefs: map[string][2]int{"N": {2, 4}}, AnimationRefs: map[string][4]int{"A": {12, 1, 0, 1}}, LampFlashRef: [2]int{40, 1}}
	c.decodeRecords(b)
	if string(c.Texts["N"]) != "NEW\x00" || c.Animations["A"].Header[0] != 3 || c.Animations["A"].Offsets[0] != 30 || c.Animations["A"].Durations[0] != 9 || c.LampFlash[0] != [4]int{5, 6, 7, 8} {
		t.Fatal("synthetic records")
	}
	c.Texts["N"][0] = 'X'
	if b[2] != 'N' {
		t.Fatal("mutable text aliases PRG")
	}
	for _, ref := range [][2]int{{-1, 1}, {59, 2}, {0, -1}} {
		t.Run(fmt.Sprint(ref), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("truncated record accepted")
				}
			}()
			recordSlice(b, ref[0], ref[1])
		})
	}
}

func TestPRGTableRecordBaselineParity(t *testing.T) {
	var proof map[string]map[string]struct {
		Count  int
		SHA256 string
	}
	raw, err := os.ReadFile("../../analysis/prg-table-record-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &proof); err != nil {
		t.Fatal(err)
	}
	packages := []string{"partyland", "speeddevils", "gameshow", "stones"}
	for n := 1; n <= 4; n++ {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			d := New(n, prg(t, n))
			file := "content.go"
			if n == 1 {
				file = "timing_data.go"
			}
			raw, err := os.ReadFile("../" + packages[n-1] + "/" + file)
			if err != nil {
				t.Fatal(err)
			}
			m := regexp.MustCompile("(?s)const \\w+ = `(.*)`").FindSubmatch(raw)
			var meta struct {
				Jingles map[string]json.RawMessage
				Lamps   map[string]struct{ Ref [2]int }
				Gates   []struct {
					Opened [4]int `json:"opened_ref"`
					Closed [4]int `json:"closed_ref"`
				}
			}
			if len(m) != 2 {
				t.Fatal("missing table layout")
			}
			if err = json.Unmarshal(m[1], &meta); err != nil {
				t.Fatal(err)
			}
			for category, p := range proof[strconv.Itoa(n)] {
				h := sha256.New()
				count := 0
				put := func(v int) { var b [8]byte; binary.LittleEndian.PutUint64(b[:], uint64(v)); h.Write(b[:]) }
				switch category {
				case "jingles":
					keys := []string{}
					for k := range meta.Jingles {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					count = len(keys)
					for _, k := range keys {
						v := d.Jingle(k)
						h.Write([]byte(k + "\x00"))
						h.Write([]byte{v.Position, v.Repeat, v.Priority})
					}
				case "lamps":
					keys := []string{}
					for k := range meta.Lamps {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					count = len(keys)
					for _, k := range keys {
						v := d.Record(meta.Lamps[k].Ref)
						h.Write([]byte(k + "\x00"))
						put(len(v))
						h.Write(v)
					}
				case "gates":
					count = len(meta.Gates)
					for _, g := range meta.Gates {
						for _, ref := range [][4]int{g.Opened, g.Closed} {
							v := d.StridedRecord(ref)
							put(len(v))
							h.Write(v)
						}
					}
				}
				if count != p.Count || hashString(h) != p.SHA256 {
					t.Fatalf("table %d %s baseline parity", n, category)
				}
			}
		})
	}
}
