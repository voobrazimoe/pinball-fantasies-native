package stones

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestAreaDecoderBoundsAndOwnership(t *testing.T) {
	data := make([]byte, 12)
	for i, v := range []uint16{1, 2, 3, 4, 9} {
		binary.LittleEndian.PutUint16(data[i*2:], v)
	}
	refs := map[string][2]int{"list": {0, 1}}
	handlers := map[uint16]string{9: "native"}
	a, b := decodeAreas(data, refs, handlers), decodeAreas(data, refs, handlers)
	a["list"][0].Rect[0] = 99
	data[0] = 88
	if b["list"][0].Rect != [4]int16{1, 2, 3, 4} || b["list"][0].Handler != "native" {
		t.Fatal("decoded lists share ownership")
	}
	for _, ref := range [][2]int{{-1, 1}, {0, -1}, {13, 0}, {0, 2}, {0, int(^uint(0) >> 1)}} {
		t.Run(fmt.Sprint(ref), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid extent accepted")
				}
			}()
			decodeAreas(data, map[string][2]int{"list": ref}, handlers)
		})
	}
	for _, mutate := range []func([]byte){func(b []byte) { b[10] = 1 }, func(b []byte) { b[8] = 10 }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid record accepted")
				}
			}()
			bad := append([]byte(nil), data...)
			mutate(bad)
			decodeAreas(bad, refs, handlers)
		}()
	}
}
