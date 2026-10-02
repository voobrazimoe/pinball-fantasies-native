package assets

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image"
)

const IntroSHA256 = "f6b5590f88949174b8f530d5b2d5b59f329d9fde24bfa98a0154d74f19f06881"

// FrontendPicture is native IFF content, never a loaded DOS executable segment.
type FrontendPicture struct {
	Width, Height int
	Indices       []byte
	Palette       []byte
}
type FrontendArt struct {
	SidebarInfo, OptionsInfo string
	Logo, Font, MonoFont     *FrontendPicture
	Tables                   [4]*FrontendPicture
	Startup                  [8]*FrontendPicture
	HighLogo, HighMono       *FrontendPicture
}

func DecodeFrontend(data []byte) (*FrontendArt, error) {
	if fmt.Sprintf("%x", sha256.Sum256(data)) != IntroSHA256 {
		return nil, fmt.Errorf("INTRO.PRG differs from inventoried build")
	}
	offsets := []int{0x1cb10, 0x6b70, 0x8b00, 0x20430, 0x24a50, 0x29410, 0x2dd40, 0x3b810, 0x42c50, 0x46930, 0x4daa0, 0x12020, 0x17f20, 0x10e60, 0xa450, 0x34d90, 0x33410}
	pics := make([]*FrontendPicture, len(offsets))
	for i, o := range offsets {
		p, e := decodeFrontendIFF(data, o)
		if e != nil {
			return nil, fmt.Errorf("INTRO IFF %x: %w", o, e)
		}
		pics[i] = p
	}
	a := &FrontendArt{Logo: pics[0], Font: pics[1], MonoFont: pics[2],
		SidebarInfo: string(data[233806:233926]), OptionsInfo: string(data[233926:234046])}
	copy(a.Tables[:], pics[3:7])
	copy(a.Startup[:], pics[7:15])
	// INTRO's linked UNPKLBM chunky branch (3b365..3b375) constructs the
	// half-bright bank before >>2 DAC conversion. The griffin uses that branch.
	// Keep the generic IFF decode/raw CMAP fixtures independent of this runtime path.
	for n := 0; n < 2; n++ {
		p := *a.Startup[n]
		p.Palette = append([]byte(nil), pics[7].Palette...)
		for i := 0; i < 96; i++ {
			v := pics[7].Palette[i]
			dac := int(v >> 2)
			p.Palette[96+i] = DACRGB(byte(dac / 2))
		}
		a.Startup[n] = &p
	}
	a.Startup[2].Palette = append([]byte(nil), a.Startup[3].Palette...)
	a.Startup[5].Palette = append([]byte(nil), a.Startup[4].Palette...)
	a.HighLogo = pics[15]
	a.HighMono = pics[16]
	return a, nil
}

func decodeFrontendIFF(data []byte, off int) (*FrontendPicture, error) {
	if off < 0 || off > len(data)-12 || string(data[off:off+4]) != "FORM" {
		return nil, fmt.Errorf("missing FORM")
	}
	kind := string(data[off+8 : off+12])
	if kind != "ILBM" && kind != "PBM " {
		return nil, fmt.Errorf("unsupported IFF kind")
	}
	end := uint64(off) + 8 + uint64(binary.BigEndian.Uint32(data[off+4:]))
	if end > uint64(len(data)) {
		return nil, fmt.Errorf("truncated FORM")
	}
	var hdr, pal, body []byte
	for pos := off + 12; pos < int(end); {
		if int(end)-pos < 8 {
			return nil, fmt.Errorf("truncated chunk")
		}
		n := uint64(binary.BigEndian.Uint32(data[pos+4:]))
		next := uint64(pos) + 8 + n + (n & 1)
		if next > end {
			return nil, fmt.Errorf("chunk exceeds FORM")
		}
		v := data[pos+8 : pos+8+int(n)]
		switch string(data[pos : pos+4]) {
		case "BMHD":
			hdr = v
		case "CMAP":
			pal = v
		case "BODY":
			body = v
		}
		pos = int(next)
	}
	if len(hdr) != 20 || hdr[10] != 1 || hdr[9] == 1 {
		return nil, fmt.Errorf("unsupported header")
	}
	w, h := int(binary.BigEndian.Uint16(hdr)), int(binary.BigEndian.Uint16(hdr[2:]))
	planes := int(hdr[8])
	if w <= 0 || w > 640 || h <= 0 || h > 512 || planes < 1 || planes > 8 || len(pal) != 3*(1<<planes) || body == nil {
		return nil, fmt.Errorf("invalid image")
	}
	p := &FrontendPicture{Width: w, Height: h, Indices: make([]byte, w*h), Palette: append([]byte(nil), pal...)}
	if kind == "PBM " {
		pixels, e := decodeByteRun1(body, (w+1)&^1, h)
		if e != nil {
			return nil, e
		}
		for y := 0; y < h; y++ {
			copy(p.Indices[y*w:(y+1)*w], pixels[y*((w+1)&^1):y*((w+1)&^1)+w])
		}
	} else {
		stride := ((w + 15) / 16) * 2
		// ByteRun1 ILBM is row/plane interleaved; runs cannot cross plane rows.
		pixels, e := decodeByteRun1(body, stride, h*planes)
		if e != nil {
			return nil, e
		}
		for y := 0; y < h; y++ {
			for plane := 0; plane < planes; plane++ {
				for x := 0; x < w; x++ {
					if pixels[(y*planes+plane)*stride+x/8]&(128>>uint(x&7)) != 0 {
						p.Indices[y*w+x] |= 1 << uint(plane)
					}
				}
			}
		}
	}
	// Original VGA palette writes retain the upper six bits of each IFF component.
	for i, v := range p.Palette {
		p.Palette[i] = DACRGB(v >> 2)
	}
	return p, nil
}
func (p *FrontendPicture) Frame() *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, p.Width, p.Height))
	for i, v := range p.Indices {
		copy(out.Pix[i*4:i*4+3], p.Palette[int(v)*3:int(v)*3+3])
		out.Pix[i*4+3] = 255
	}
	return out
}
