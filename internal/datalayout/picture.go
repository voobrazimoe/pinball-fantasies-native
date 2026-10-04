package datalayout

import (
	"encoding/binary"
	"fmt"
)

// Picture anchors identify the linked layout before a profile is selected.
// Full row/plane decompression remains in the common assets decoder.
func validatePicture(data []byte, p Picture) error {
	at := p.Offset
	if at < 0 || at > len(data)-12 || string(data[at:at+4]) != "FORM" || string(data[at+8:at+12]) != p.Kind {
		return fmt.Errorf("missing FORM/%s anchor at %#x", p.Kind, at)
	}
	size := uint64(binary.BigEndian.Uint32(data[at+4:]))
	end := uint64(at) + 8 + size
	if size < 4 || end > uint64(len(data)) {
		return fmt.Errorf("truncated FORM at %#x", at)
	}
	var hdr, pal, body []byte
	for pos := at + 12; pos < int(end); {
		if int(end)-pos < 8 {
			return fmt.Errorf("truncated IFF chunk")
		}
		n := uint64(binary.BigEndian.Uint32(data[pos+4:]))
		next := uint64(pos) + 8 + n + (n & 1)
		if next > end {
			return fmt.Errorf("IFF chunk exceeds FORM")
		}
		value := data[pos+8 : pos+8+int(n)]
		switch string(data[pos : pos+4]) {
		case "BMHD":
			if hdr != nil {
				return fmt.Errorf("duplicate BMHD")
			}
			hdr = value
		case "CMAP":
			if pal != nil {
				return fmt.Errorf("duplicate CMAP")
			}
			pal = value
		case "BODY":
			if body != nil {
				return fmt.Errorf("duplicate BODY")
			}
			body = value
		}
		pos = int(next)
	}
	if len(hdr) != 20 || int(binary.BigEndian.Uint16(hdr)) != p.Width || int(binary.BigEndian.Uint16(hdr[2:])) != p.Height || int(hdr[8]) != p.Planes || (hdr[9] != 0 && hdr[9] != 2) || hdr[10] != 1 {
		return fmt.Errorf("incompatible IFF geometry/header at %#x", at)
	}
	if len(pal) != 3*(1<<p.Planes) || len(body) == 0 {
		return fmt.Errorf("missing compatible CMAP/BODY at %#x", at)
	}
	return nil
}
