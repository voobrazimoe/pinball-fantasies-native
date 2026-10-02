// Package audio decodes only the supplied Party Land module. It has no host clock.
package audio

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const ModuleSHA256 = "a0877e4372abe64b70d9e361bf257ea5a84c948771f0eace3433d5f6399060b5"
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

func Decode(data []byte) (*Module, error) {
	if fmt.Sprintf("%x", sha256.Sum256(data)) != ModuleSHA256 {
		return nil, fmt.Errorf("TABLE1.MOD differs from supplied Party Land module")
	}
	return decodeModule(data, false)
}

// DecodeSpeedDevils uses the same native tracker and validates TABLE2.MOD.
func DecodeSpeedDevils(data []byte) (*Module, error) {
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "728629c54311386781271308e181ac0435f0582e90870accff0a42270d467529" {
		return nil, fmt.Errorf("TABLE2.MOD differs from supplied Speed Devils module")
	}
	return decodeModule(data, false)
}

// DecodeGameshow validates original TABLE3.MOD and reuses the native tracker.
func DecodeGameshow(data []byte) (*Module, error) {
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "fb7bfd1c96a462cb03999d2e6f843a20d3de69ba05fcbd384a9f1c131b9a563a" {
		return nil, fmt.Errorf("TABLE3.MOD differs from supplied Gameshow module")
	}
	return decodeModuleWithOrders(data, false, 63)
}

// DecodeFrontend accepts only the two inventoried original frontend modules.
func DecodeFrontend(data []byte) (*Module, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	if hash != "9ecb5813ba1a5f606b47f0dbb9cc65c5dc5c0f07596ccd23fdf019b0283f1dcc" && hash != "aa5003c275b494062f37f44e8c77105b8a420555f4bd6ff53d7698f89c540f21" {
		return nil, fmt.Errorf("frontend MOD differs from inventoried build")
	}
	return decodeModule(data, hash == "9ecb5813ba1a5f606b47f0dbb9cc65c5dc5c0f07596ccd23fdf019b0283f1dcc")
}

func decodeModule(data []byte, omittedSilentSamples bool) (*Module, error) {
	return decodeModuleWithOrders(data, omittedSilentSamples, int(data[950]))
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
func DecodeStones(data []byte) (*Module, error) {
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "31ad7e671ae77c07c3d075e2f1fecd3d918fd921fa23acd9a1b0b6fc07fbbcea" {
		return nil, fmt.Errorf("TABLE4.MOD differs from supplied Stones module")
	}
	return decodeModule(data, false)
}
