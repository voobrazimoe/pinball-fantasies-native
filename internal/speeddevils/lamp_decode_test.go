package speeddevils

import (
	"testing"
)

func TestLampDecoderBoundsAndOwnership(t *testing.T) {
	data := []byte{254, 2, 1, 2, 3, 4, 5, 6}
	a := decodeLamp(data, 0)
	b := decodeLamp(data, 0)
	a.rgb[0] = 99
	data[2] = 88
	if b.start != 254 || b.rgb[0] != 1 {
		t.Fatal("lamp packets share ownership")
	}
	for _, q := range []struct {
		data []byte
		at   int
	}{{data, -1}, {data, 9}, {data, 7}, {[]byte{255, 2, 1, 2, 3, 4, 5, 6}, 0}, {[]byte{1, 1, 2}, 0}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid packet accepted")
				}
			}()
			decodeLamp(q.data, q.at)
		}()
	}
}
