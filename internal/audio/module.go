// Package audio decodes supported four-channel game modules. It has no host clock.
package audio

import (
	"encoding/binary"
	"fmt"
)

const Rate = 48000

type Sample struct {
	Name                  string
	PCM                   []int8
	Volume, Fine          int
	LoopStart, LoopLength int
}
type Note struct{ Sample, Period, Effect, Param int }
type Module struct {
	Orders   []int
	Patterns [][64][4]Note
	Samples  [31]Sample
}

// Table profiles retain source-referenced order extents; PCM bytes are not an oracle.
func Decode(data []byte) (*Module, error)            { return compatibleModule(data, false, 64) }
func DecodeSpeedDevils(data []byte) (*Module, error) { return compatibleModule(data, false, 65) }
func DecodeGameshow(data []byte) (*Module, error)    { return compatibleModule(data, false, 63) }
func DecodeIntro(data []byte) (*Module, error)       { return compatibleModule(data, true, 44) }
func DecodeMenu(data []byte) (*Module, error)        { return compatibleModule(data, false, 15) }

// DecodeFrontend preserves the library API. Host loading uses explicit roles.
func DecodeFrontend(data []byte) (*Module, error) {
	if len(data) < 1084 {
		return nil, fmt.Errorf("unsupported MOD layout: truncated header")
	}
	if data[950] >= 44 {
		return DecodeIntro(data)
	}
	return DecodeMenu(data)
}

func compatibleModule(data []byte, intro bool, requiredOrders int) (*Module, error) {
	if len(data) < 1084 || string(data[1080:1084]) != "M.K." {
		return nil, fmt.Errorf("unsupported MOD layout: require 31-sample four-channel M.K. module")
	}
	count := int(data[950])
	if count < 1 || count > 128 {
		return nil, fmt.Errorf("invalid MOD song length")
	}
	// SHOW references orders 61/62 beyond its legacy length byte.
	if count < requiredOrders {
		if requiredOrders != 63 || count != 61 {
			return nil, fmt.Errorf("unsupported MOD layout: need %d source-referenced orders", requiredOrders)
		}
		count = requiredOrders
	}
	maxPattern := 0
	for _, p := range data[952 : 952+count] {
		if p > 127 {
			return nil, fmt.Errorf("invalid MOD pattern reference")
		}
		if int(p) > maxPattern {
			maxPattern = int(p)
		}
	}
	off := 1084 + (maxPattern+1)*1024
	if off > len(data) {
		return nil, fmt.Errorf("truncated MOD patterns")
	}
	used := [31]bool{}
	for q := 1084; q < off; q += 4 {
		sample := int(data[q]&240) | int(data[q+2]>>4)
		if sample > 31 {
			return nil, fmt.Errorf("invalid MOD sample number")
		}
		if sample > 0 {
			used[sample-1] = true
		}
		if data[q+2]&15 == 15 && (data[q+3] == 0 || data[q+3] > 31) {
			return nil, fmt.Errorf("unsupported MOD layout: tracker requires speed-only F01..F1F")
		}
		if data[q+2]&15 == 11 && int(data[q+3]) >= count {
			return nil, fmt.Errorf("MOD jump outside orders")
		}
		if data[q+2]&15 == 13 && (data[q+3]>>4 > 6 || data[q+3]&15 > 9 || int(data[q+3]>>4)*10+int(data[q+3]&15) > 63) {
			return nil, fmt.Errorf("MOD break outside rows")
		}
	}
	total := 0
	for i := 0; i < 31; i++ {
		h := data[20+i*30 : 50+i*30]
		n := int(binary.BigEndian.Uint16(h[22:])) * 2
		loopStart, loopLength := int(binary.BigEndian.Uint16(h[26:]))*2, int(binary.BigEndian.Uint16(h[28:]))*2
		if h[24] > 15 || h[25] > 64 || (loopLength > 2 && (loopStart > n || loopLength > n-loopStart)) {
			return nil, fmt.Errorf("invalid MOD sample %d metadata", i+1)
		}
		if used[i] && n == 0 {
			return nil, fmt.Errorf("missing MOD pattern sample %d", i+1)
		}
		total += n
	}
	omitted := false
	// INTRO's verified omitted slots have two-byte silent headers, no loop,
	// and no pattern references. Accept both full storage and legacy omission.
	if intro && len(data)-off == total-22+2 {
		omitted = true
		for i := 20; i < 31; i++ {
			h := data[20+i*30 : 50+i*30]
			if used[i] || binary.BigEndian.Uint16(h[22:]) != 1 || h[25] != 0 || binary.BigEndian.Uint16(h[28:]) > 1 {
				return nil, fmt.Errorf("unsupported INTRO omitted sample layout")
			}
		}
	} else if len(data)-off != total && len(data)-off != total+2 {
		return nil, fmt.Errorf("unsupported MOD layout: sample extent differs (truncated or extra patterns)")
	}
	m, err := decodeModuleWithOrders(data, omitted, count)
	if err != nil {
		return nil, err
	}
	if requiredOrders >= 63 {
		// All table effects address these slots directly; other slots are checked
		// by pattern references. Samples may change, but cannot disappear.
		required := map[int][]int{64: {7, 22, 23, 24, 25, 28, 29, 30}, 65: {23, 24, 25, 26, 27, 28, 29, 30}, 63: {7, 15, 22, 23, 24, 25, 28, 29, 30}, 66: {2, 6, 10, 23, 24, 25, 28, 29, 30}}
		for _, sample := range required[requiredOrders] {
			if len(m.Samples[sample-1].PCM) == 0 {
				return nil, fmt.Errorf("missing table effect sample %d", sample)
			}
		}
	}
	return m, nil
}

// Gameshow's source explicitly references the populated order 62 beyond its
// 61-order song header. Decode its verified 63-entry extent, including all
// 64 stored patterns before samples. Other modules keep their accepted extent.
func decodeModuleWithOrders(data []byte, omittedSilentSamples bool, orderCount int) (*Module, error) {
	m := &Module{}
	maxPattern := 0
	for _, v := range data[952 : 952+orderCount] {
		m.Orders = append(m.Orders, int(v))
		if int(v) > maxPattern {
			maxPattern = int(v)
		}
	}
	m.Patterns = make([][64][4]Note, maxPattern+1)
	for p := range m.Patterns {
		for r := 0; r < 64; r++ {
			for c := 0; c < 4; c++ {
				o := 1084 + p*1024 + r*16 + c*4
				q := data[o : o+4]
				m.Patterns[p][r][c] = Note{int(q[0]&240) | int(q[2]>>4), int(q[0]&15)<<8 | int(q[1]), int(q[2] & 15), int(q[3])}
			}
		}
	}
	off := 1084 + len(m.Patterns)*1024
	for i := range m.Samples {
		h := data[20+i*30 : 50+i*30]
		s := &m.Samples[i]
		s.Name = string(h[:22])
		n := int(binary.BigEndian.Uint16(h[22:])) * 2
		s.Volume = int(h[25])
		s.Fine = int(h[24] & 15)
		if s.Fine >= 8 {
			s.Fine -= 16
		}
		s.LoopStart = int(binary.BigEndian.Uint16(h[26:])) * 2
		s.LoopLength = int(binary.BigEndian.Uint16(h[28:])) * 2
		s.PCM = make([]int8, n)
		// Supplied INTRO.MOD omits unused silent slots 21..31 and appends
		// the two-byte INTRO checksum instead. No playable sample is truncated.
		if omittedSilentSamples && i >= 20 && n == 2 {
			continue
		}
		if n > len(data)-off {
			return nil, fmt.Errorf("truncated MOD sample %d", i+1)
		}
		for j := range s.PCM {
			s.PCM[j] = int8(data[off+j])
		}
		off += n
	}
	return m, nil
}

// PLAND.ASM SOUND STRUCTURES: sample (one based), note index (zero based), channel 3.
type Effect struct{ Sample, Note int }

var Effects = map[string]Effect{
	"SBRICKNER": {22, 18}, "SBRICKUPP": {23, 23}, "SBUMPER1": {24, 25}, "SBUMPER2": {24, 23}, "SBUMPER3": {24, 21},
	"SFLIPPUPP": {25, 22}, "SRINNER": {28, 18}, "SNEWBALL": {28, 18}, "SKICKER": {29, 18}, "SFJADER": {30, 18}, "SGROP": {23, 23},
	"SBYGEL1": {7, 10}, "SBYGEL2": {7, 12}, "SBYGEL3": {7, 8}, "SBYGEL4": {7, 17}, "SBYGEL5": {7, 8},
	"S_TOUCH1": {7, 14}, "S_TOUCH2": {7, 10}, "S_SCORELJUD": {7, 18}, "S_ADDPLAYER2": {7, 18},
}

// DecodeStones uses the physically verified 66-order/64-pattern TABLE4 module.
func DecodeStones(data []byte) (*Module, error) { return compatibleModule(data, false, 66) }
