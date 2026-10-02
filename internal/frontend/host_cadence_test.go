package frontend

import (
	"bytes"
	"crypto/sha256"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/settings"
	"testing"
)

// A slow host presents one frame for several original syncs. PCM must still be
// generated for every source tick, in order, independent of frame consumption.
func TestHostFrameCadenceLeavesTicksAndPCMIdentical(t *testing.T) {
	a, b := configured(t, settings.Legacy()), configured(t, settings.Legacy())
	for _, r := range []*Runtime{a, b} {
		for _, key := range []Key{Space, F1, F1} {
			runKey(t, r, key)
		}
	}
	for tick := 0; tick < 1200; tick++ {
		if tick == 100 || tick == 700 {
			a.Model.Session.Release(32, 0)
			b.Model.Session.Release(32, 0)
		}
		in := Input{Left: tick%93 < 18, Right: tick%71 < 14}
		for _, r := range []*Runtime{a, b} {
			if e := r.Update(in); e != nil {
				t.Fatal(e)
			}
		}
		if !bytes.Equal(a.PCM, b.PCM) {
			t.Fatalf("PCM changed with frame cadence at source tick %d", tick)
		}
		a.Frame()
		if tick%8 == 7 {
			b.Frame()
		}
	}
	for _, r := range []*Runtime{a, b} {
		g := r.Model.Session.(*partyland.Game)
		if g.Tick != 1200 || g.Score.Uint64() != 2_300_000 || g.BallNumber != 2 {
			t.Fatal("source oracle changed", g.Tick, g.Score, g.BallNumber)
		}
	}
	if ah, bh := sha256.Sum256(a.Frame().Pix), sha256.Sum256(b.Frame().Pix); ah != bh {
		t.Fatalf("logical frame changed: %x != %x", ah, bh)
	}
}
