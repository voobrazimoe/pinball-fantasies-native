//go:build dmoimpl1 || demodev

package partyland

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"pinballfantasies/internal/presentation"

	"reflect"
)

// Normal completion is a value, never an error or a process exit.
type demoTerminal struct {
	Status      uint16
	Calculation uint16
	Visits      int
}

func (r demoTerminal) String() string {
	return fmt.Sprintf("DEMO_QUIT_REACHED status=%d calculation=%d visits=%d", r.Status, r.Calculation, r.Visits)
}

type demoFade struct {
	Active            bool
	Palette, Snapshot [768]byte // owned VGA DAC units, not shared host RGB memory
	Volume            []uint16  // actual AL=6 requests, CX=remaining, every sixteen visits
}

// matrixConsumer extends only the two unavailable commands. Dispatch, cursor,
// timing, NEXT_A, scroll, printing, waits and flash remain the shared interpreter.
func (d *demoCore) dispatch(c matrixCommand) bool {
	if d.dispatchScored(c) {
		return true
	}
	switch c.Op {
	case "_DEMO_PARTY_FLASHOFF":
		d.game.Display.KillFlash()
		d.game.matrix.remaining = 1
		return true
	case "_DEMO_FADE":
		d.game.matrix.remaining = uint16(c.Num(0))
		d.fade.Active = true
		d.fade.Snapshot = d.fade.Palette
		return true
	case "_DEMO_QUIT":
		d.terminal = &demoTerminal{uint16(c.Num(0)), d.counter, d.visits}
		d.game.matrix.active = false
		return true
	}
	return false
}
func (d *demoCore) step(op string, left *uint16) bool {
	if op != "_DEMO_FADE" {
		return false
	}
	*left--
	for i, v := range d.fade.Snapshot {
		d.fade.Palette[i] = byte(uint16(v) * (*left) >> 8)
	}
	if *left&15 == 0 {
		d.fade.Volume = append(d.fade.Volume, *left)
	}
	return true

}
func (d *demoCore) matrixCompletes() bool {
	m := d.game.matrix
	switch m.op {
	case "_PARTYON":
		return false
	case "_SCROLL":
		return d.game.Display.ScrollCompletes(m.textLeft)
	default:
		return m.remaining == 1
	}
}
func (d *demoCore) expiryCommand(c presentation.Command) bool {
	if !d.expiryPresentation {
		return false
	}
	if c.Op == "_SCROLL" {
		if len(c.Args) != 1 {
			return false
		}
		index := 0
		if c.Args[0] == "DEMO_EXPIRY_TEXT2" {
			index = 1
		}
		if len(d.expiryTexts[index]) == 0 || !bytes.Equal(d.game.Display.Content.Texts[c.Args[0]], d.expiryTexts[index]) {
			return false
		}
	}
	if c.Op == "_PRINT13_NUMBER" {
		if font, ok := d.game.Display.Content.Fonts["13"]; !ok || font != d.expiryFont {
			return false
		}
	}
	for _, want := range demoExpiryProgram() {
		if c.Op == want.Op && reflect.DeepEqual(c.Args, want.Args) && sameDemoNums(c.Nums, want.Nums) {
			return true
		}
	}
	return false
}

// Caller first pins the entire private demo installation. This validates only
// the reviewed linked stream and the concrete presentation consumers, no graph.
func (d *demoCore) loadExpiry(demo, canonical []byte) error {
	fail := func() error {
		return d.reject("expiry presentation", "linked operands/glyphs", "private consumer correspondence unavailable")
	}
	sites := []int{0x1ba17, 0x1ba19, 0x1ba1d, 0x1ba21, 0x1ba27, 0x1ba2b, 0x1ba2f, 0x1ba33, 0x1ba37, 0x1ba3b}
	args := [][]uint16{nil, {7313}, {1}, {18101, 344}, {100}, {1}, {7412}, {256}, {100}, {0}}

	handlers := []uint16{0x4ebe, 0x456a, 0x459b, 0x466b, 0x4da3, 0x4d3c, 0x456a, 0x4f56, 0x4da3, 0x39db}
	if len(demo) < 0x1ba41 {
		return fail()
	}
	for i, at := range sites {
		if binary.LittleEndian.Uint16(demo[at:]) != handlers[i] {
			return fail()
		}
		for j, v := range args[i] {
			if binary.LittleEndian.Uint16(demo[at+2+j*2:]) != v {
				return fail()
			}
		}
	}

	if binary.LittleEndian.Uint16(demo[sites[7]:]) != 0x4f56 || binary.LittleEndian.Uint16(demo[sites[9]:]) != 0x39db {
		return fail()
	}
	if err := d.loadPartyPresentation(demo, canonical); err != nil {
		return err
	}
	for i, at := range []int{7313, 7412} {
		length := []int{98, 88}[i]
		text := demo[0x19db0+at : 0x19db0+at+length+1]
		if text[length] != 255 || bytes.IndexByte(text, 255) != length {
			return fail()
		}
		for _, ch := range text[:length] {
			for _, layout := range [][2]int{{0x5e00, 0x7f00}, {0x6000, 0x7390}} {
				a := layout[1] + int(binary.LittleEndian.Uint16(canonical[0x19d40+layout[0]+2*int(ch):]))
				b := layout[1] + 0x70 + int(binary.LittleEndian.Uint16(demo[0x19db0+layout[0]+0x100+2*int(ch):]))
				for n := 0; n < 1024; n++ {
					if a >= len(canonical) || b >= len(demo) || canonical[a] != demo[b] {
						return fail()
					}
					if canonical[a] == 0xc3 {
						break
					}
					if a+4 > len(canonical) || b+4 > len(demo) || !bytes.Equal(canonical[a:a+4], demo[b:b+4]) || canonical[a] != 0x88 || (canonical[a+1] != 0x87 && canonical[a+1] != 0xa7) {
						return fail()
					}
					a += 4
					b += 4
					if n == 1023 {
						return fail()
					}
				}
			}
		}
		d.game.Display.Content.Texts[fmt.Sprintf("DEMO_EXPIRY_TEXT%d", i+1)] = append([]byte(nil), text...)
		d.expiryTexts[i] = append([]byte(nil), text...)
	}

	rgb := d.game.Display.SourceMatrixPalette(d.game.lampPalette)
	for i, v := range rgb {
		d.fade.Palette[i] = v >> 2
	}
	d.expiryFont = d.game.Display.Content.Fonts["13"]
	d.expiryPresentation = true
	return nil
}

func sameDemoNums(a, b map[int]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}
