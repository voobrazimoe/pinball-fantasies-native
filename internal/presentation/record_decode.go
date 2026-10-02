package presentation

import (
	"encoding/binary"
	"pinballfantasies/internal/tablelogic"
	"strings"
)

// recordSlice never accepts truncated/present input. A nil PRG is permitted for
// metadata-only structural tests; it does not provide synthetic game content.
func recordSlice(data []byte, at, size int) []byte {
	if at < 0 || size < 0 || at > len(data) || size > len(data)-at {
		panic("PRG content record outside supplied input")
	}
	return data[at : at+size]
}

func (c *Content) decodeRecords(data []byte) {
	for label, ref := range c.TextRefs {
		if len(data) == 0 {
			c.Texts[label] = nil
		} else {
			c.Texts[label] = append([]byte{}, recordSlice(data, ref[0], ref[1])...)
		}
	}
	c.Animations = make(map[string]Animation, len(c.AnimationRefs))
	for label, ref := range c.AnimationRefs {
		if len(data) == 0 {
			c.Animations[label] = Animation{}
			continue
		}
		at, count, base, prefix := ref[0], ref[1], ref[2], ref[3]
		if count < 0 || count > len(data)/4 {
			panic("invalid PRG animation frame count")
		}
		raw := recordSlice(data, at, 4+4*count)
		a := Animation{Header: []uint16{0, binary.LittleEndian.Uint16(raw), binary.LittleEndian.Uint16(raw[2:])}}
		if prefix != 0 {
			a.Header[0] = binary.LittleEndian.Uint16(recordSlice(data, at-2, 2))
		}
		for i := 0; i < count; i++ {
			p := raw[4+i*4:]
			off := base + int(binary.LittleEndian.Uint16(p))
			recordSlice(data, off, 2)
			a.Offsets = append(a.Offsets, off)
			a.Durations = append(a.Durations, binary.LittleEndian.Uint16(p[2:]))
		}
		c.Animations[label] = a
	}
	if len(data) == 0 {
		return
	}
	at, count := c.LampFlashRef[0], c.LampFlashRef[1]
	if count < 0 || count > len(data)/10 {
		panic("invalid PRG lamp flash count")
	}
	raw := recordSlice(data, at, count*10)
	c.LampFlash = make([][4]int, count)
	for i := range c.LampFlash {
		for j := range c.LampFlash[i] {
			c.LampFlash[i][j] = int(binary.LittleEndian.Uint16(raw[i*10+2+j*2:]))
		}
	}
}

// Record returns an owned copy of a bounded literal PRG record.
func (d *Display) Record(ref [2]int) []byte {
	return append([]byte(nil), recordSlice(d.data, ref[0], ref[1])...)
}

// Jingle interprets the three-byte tracker control record, not machine code.
func (d *Display) Jingle(label string) tablelogic.JingleSpec {
	b, ok := d.Content.Texts[strings.ToUpper(label)]
	if !ok || len(b) != 3 {
		panic("missing PRG jingle record " + label)
	}
	return tablelogic.JingleSpec{Position: b[0], Repeat: b[1], Priority: b[2]}
}

// StridedRecord reads interleaved gate mask rows as structured data.
func (d *Display) StridedRecord(ref [4]int) []byte {
	at, width, count, stride := ref[0], ref[1], ref[2], ref[3]
	if width < 0 || count < 0 || stride < width || width > len(d.data) || count > len(d.data) {
		panic("invalid PRG strided record")
	}
	var out []byte
	for i := 0; i < count; i++ {
		out = append(out, recordSlice(d.data, at, width)...)
		if stride > len(d.data)-at && i+1 < count {
			panic("PRG stride outside input")
		}
		at += stride
	}
	return out
}
