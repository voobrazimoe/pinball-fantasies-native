// Package assets decodes original content as data; no executable code is loaded.
package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image"
	"os"
)

const Width = 320
const Height = 576
const Table1SHA256 = "4d7a69e7dc95260ad2541c6981a11ab842e2f1f20e45447e5613688b86e38414"

// Original: FANTASIE.ASM, INIT_GFX: STAGE1_1..4 at y=0,144,288,432
// (after the separate SPLH score panel); PLAND.ASM, BANH=576.
// Offsets are from static IFF chunk inspection of the pinned TABLE1.PRG.
var stripOffsets = [...]int{336944, 366176, 399776, 437168}

type Playfield struct {
	Indices []byte
	Palette [768]byte
}

func LoadPartyLand(path string) (*Playfield, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return DecodePartyLand(data)
}

// DecodePartyLand intentionally accepts only the inventoried installation build.
// MZ headers and machine instructions are never interpreted or executed.
func DecodePartyLand(data []byte) (*Playfield, error) {
	if fmt.Sprintf("%x", sha256.Sum256(data)) != Table1SHA256 {
		return nil, fmt.Errorf("TABLE1.PRG differs from the inventoried build (want SHA-256 %s)", Table1SHA256)
	}
	field := &Playfield{Indices: make([]byte, 0, Width*Height)}
	for _, off := range stripOffsets {
		indices, palette, err := decodeStrip(data, off)
		if err != nil {
			return nil, fmt.Errorf("strip at %d: %w", off, err)
		}
		field.Indices = append(field.Indices, indices...)
		// Original INIT_GFX copies the palette from STAGE1_4 into PALLE.
		copy(field.Palette[:], palette)
	}
	return field, nil
}

func decodeStrip(data []byte, offset int) ([]byte, []byte, error) {
	return decodeStripHeight(data, offset, 144)
}

func decodeStripHeight(data []byte, offset, height int) ([]byte, []byte, error) {
	if offset < 0 || offset > len(data)-12 || string(data[offset:offset+4]) != "FORM" || string(data[offset+8:offset+12]) != "PBM " {
		return nil, nil, fmt.Errorf("missing FORM/PBM")
	}
	size := uint64(binary.BigEndian.Uint32(data[offset+4:]))
	end64 := uint64(offset) + 8 + size
	if size < 4 || end64 > uint64(len(data)) {
		return nil, nil, fmt.Errorf("invalid FORM size")
	}
	end := int(end64)
	var header, palette, body []byte
	for pos := offset + 12; pos < end; {
		if end-pos < 8 {
			return nil, nil, fmt.Errorf("truncated chunk header")
		}
		n := uint64(binary.BigEndian.Uint32(data[pos+4:]))
		next := uint64(pos) + 8 + n + (n & 1)
		if next > uint64(end) {
			return nil, nil, fmt.Errorf("chunk exceeds FORM")
		}
		payload := data[pos+8 : pos+8+int(n)]
		switch string(data[pos : pos+4]) {
		case "BMHD":
			if header != nil {
				return nil, nil, fmt.Errorf("duplicate BMHD")
			}
			header = payload
		case "CMAP":
			if palette != nil {
				return nil, nil, fmt.Errorf("duplicate CMAP")
			}
			palette = payload
		case "BODY":
			if body != nil {
				return nil, nil, fmt.Errorf("duplicate BODY")
			}
			body = payload
		}
		pos = int(next)
	}
	if len(header) != 20 || binary.BigEndian.Uint16(header) != Width || int(binary.BigEndian.Uint16(header[2:])) != height || header[8] != 8 || (header[9] != 0 && header[9] != 2) || header[10] != 1 {
		return nil, nil, fmt.Errorf("unsupported BMHD (want 320x144, 8-bit, no mask plane, ByteRun1)")
	}
	if len(palette) != 768 || body == nil {
		return nil, nil, fmt.Errorf("missing 256-color CMAP or BODY")
	}
	indices, err := decodeByteRun1(body, Width, height)
	return indices, palette, err
}

// PBM ByteRun1 expands each chunky, even-padded scanline independently.
// 0..127 = literal count+1; -1..-127 = repeat 1-count; -128 = no-op.
func decodeByteRun1(src []byte, width, height int) ([]byte, error) {
	out := make([]byte, 0, width*height)
	p := 0
	for y := 0; y < height; y++ {
		rowEnd := len(out) + width
		for len(out) < rowEnd {
			if p >= len(src) {
				return nil, fmt.Errorf("truncated row %d", y)
			}
			n := int(int8(src[p]))
			p++
			if n == -128 {
				continue
			}
			count := n + 1
			if n < 0 {
				count = 1 - n
			}
			if count > rowEnd-len(out) {
				return nil, fmt.Errorf("run exceeds row %d", y)
			}
			if n >= 0 {
				if count > len(src)-p {
					return nil, fmt.Errorf("truncated literal")
				}
				out = append(out, src[p:p+count]...)
				p += count
			} else {
				if p >= len(src) {
					return nil, fmt.Errorf("truncated repeat")
				}
				out = append(out, bytes.Repeat(src[p:p+1], count)...)
				p++
			}
		}
	}
	if p != len(src) {
		return nil, fmt.Errorf("%d trailing BODY bytes", len(src)-p)
	}
	return out, nil
}

// Framebuffer preserves the original PBM RGB bytes. It does not simulate VGA
// DAC quantization, palette flashing, flipper overlays, or the score panel.
func (p *Playfield) Framebuffer() *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, Width, Height))
	for i, c := range p.Indices {
		copy(dst.Pix[i*4:i*4+3], p.Palette[int(c)*3:int(c)*3+3])
		dst.Pix[i*4+3] = 255
	}
	return dst
}
