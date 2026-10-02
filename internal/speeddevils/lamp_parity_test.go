package speeddevils

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestOriginalLampParityHash(t *testing.T) {
	g := testGame(t)
	h := sha256.New()
	for n := 1; n <= 67; n++ {
		c := g.content[n]
		h.Write([]byte{byte(n), byte(c.start), byte(len(c.rgb) / 3)})
		h.Write(c.rgb)
	}
	if fmt.Sprintf("%x", h.Sum(nil)) != "44d7f25dd7450575b6d023833825befe463d2caea7bacd54f474450b1a09fe85" {
		t.Fatal("original lamp header/packet ordering changed")
	}
}
