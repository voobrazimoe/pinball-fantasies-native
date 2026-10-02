package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"pinballfantasies/internal/testinputs"
	"testing"
)

// Expected hashes come from tools/reference_extract.py, an independent Python
// ByteRun1 extractor reading the original file, not from this Go decoder.
func TestPartyLandReference(t *testing.T) {
	fixtureBytes, err := os.ReadFile("../../analysis/pf1-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		FileSHA    string `json:"file_sha256"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		IndicesSHA string `json:"indices_sha256"`
		RGBASHA    string `json:"rgba_sha256"`
		Strips     []struct {
			Offset     int    `json:"offset"`
			Size       int    `json:"form_size"`
			BodySize   int    `json:"body_size"`
			IndicesSHA string `json:"indices_sha256"`
			PaletteSHA string `json:"palette_sha256"`
		} `json:"strips"`
	}
	if err = json.Unmarshal(fixtureBytes, &fixture); err != nil {
		t.Fatal(err)
	}
	testinputs.Require(t, "../../TABLE1.PRG")
	data, err := os.ReadFile("../../TABLE1.PRG")
	if err != nil {
		t.Fatal(err)
	}
	hash := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	if hash(data) != fixture.FileSHA || fixture.FileSHA != Table1SHA256 {
		t.Fatal("original file hash mismatch")
	}
	for _, s := range fixture.Strips {
		if int(binary.BigEndian.Uint32(data[s.Offset+4:]))+8 != s.Size {
			t.Fatal("FORM size mismatch")
		}
		idx, pal, err := decodeStrip(data, s.Offset)
		if err != nil {
			t.Fatal(err)
		}
		if len(idx) != 320*144 || hash(idx) != s.IndicesSHA || hash(pal) != s.PaletteSHA {
			t.Fatalf("strip %d mismatch", s.Offset)
		}
	}
	field, err := DecodePartyLand(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(field.Indices) != fixture.Width*fixture.Height || Width != fixture.Width || Height != fixture.Height {
		t.Fatal("dimensions")
	}
	if hash(field.Indices) != fixture.IndicesSHA {
		t.Fatal("indices differ from independent extraction")
	}
	frame := field.Framebuffer()
	if hash(frame.Pix) != fixture.RGBASHA {
		t.Fatal("framebuffer differs from independent extraction")
	}
	second, err := DecodePartyLand(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(frame.Pix, second.Framebuffer().Pix) {
		t.Fatal("repeated render differs")
	}
	data[len(data)-1] ^= 1
	if _, err := DecodePartyLand(data); err == nil {
		t.Fatal("modified file accepted")
	}
}

func TestByteRun1(t *testing.T) {
	// Hand-authored vectors cover all three signed-control meanings.
	for _, tc := range []struct {
		name string
		src  []byte
		w, h int
		want []byte
		bad  bool
	}{
		{"literal", []byte{2, 1, 2, 3}, 3, 1, []byte{1, 2, 3}, false},
		{"repeat", []byte{254, 7}, 3, 1, []byte{7, 7, 7}, false},
		{"noop", []byte{128, 0, 9}, 1, 1, []byte{9}, false},
		{"rows", []byte{255, 4, 1, 5, 6}, 2, 2, []byte{4, 4, 5, 6}, false},
		{"truncated literal", []byte{2, 1}, 3, 1, nil, true},
		{"truncated repeat", []byte{254}, 3, 1, nil, true},
		{"crosses row", []byte{254, 4}, 2, 1, nil, true},
		{"missing row", []byte{255, 4}, 2, 2, nil, true},
		{"trailing", []byte{0, 1, 128}, 1, 1, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeByteRun1(tc.src, tc.w, tc.h)
			if (err != nil) != tc.bad || !bytes.Equal(got, tc.want) {
				t.Fatalf("got %v, %v", got, err)
			}
		})
	}
}

func TestStripBounds(t *testing.T) {
	testinputs.Require(t, "../../TABLE1.PRG")
	original, err := os.ReadFile("../../TABLE1.PRG")
	if err != nil {
		t.Fatal(err)
	}
	base := stripOffsets[0]
	for _, n := range []int{0, base, base + 11, base + 12, base + 100} {
		if _, _, err := decodeStrip(original[:n], base); err == nil {
			t.Fatalf("accepted truncated size %d", n)
		}
	}
	for _, off := range []int{-1, len(original), int(^uint(0) >> 1)} {
		if _, _, err := decodeStrip(original, off); err == nil {
			t.Fatal("invalid offset accepted")
		}
	}
	broken := append([]byte(nil), original...)
	binary.BigEndian.PutUint32(broken[base+4:], 0xffffffff)
	if _, _, err := decodeStrip(broken, base); err == nil {
		t.Fatal("oversized FORM accepted")
	}
	broken = append([]byte(nil), original...)
	binary.BigEndian.PutUint32(broken[base+16:], 0xffffffff)
	if _, _, err := decodeStrip(broken, base); err == nil {
		t.Fatal("oversized chunk accepted")
	}
}
