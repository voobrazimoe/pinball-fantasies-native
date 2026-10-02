package stones

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"testing"
)

func TestOriginalAreaParityHash(t *testing.T) {
	g := game(t)
	names := make([]string, 0, len(g.areas))
	for name := range g.areas {
		names = append(names, name)
	}
	sort.Strings(names)
	h := sha256.New()
	count := 0
	for _, name := range names {
		h.Write(append([]byte(name), 0))
		for _, r := range g.areas[name] {
			var rect [8]byte
			for i, v := range r.Rect {
				binary.LittleEndian.PutUint16(rect[i*2:], uint16(v))
			}
			h.Write(rect[:])
			h.Write(append([]byte(r.Handler), 0))
			count++
		}
	}
	if count != 44 || fmt.Sprintf("%x", h.Sum(nil)) != "5cced501ceda5c34c378aad24ce601f1a887bdd2736ac1ef157967e5c6e3966a" {
		t.Fatal("original area/handler ordering changed", count)
	}
}
