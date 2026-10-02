package presentation

import (
	"bytes"
	"image"
	"testing"
)

func TestFullTableExactRowIdentity(t *testing.T) {
	field := image.NewRGBA(image.Rect(0, 0, 320, 576))
	for y := 0; y < 576; y++ {
		for x := 0; x < 320; x++ {
			o := field.PixOffset(x, y)
			copy(field.Pix[o:o+4], []byte{byte(y), byte(y >> 8), byte(x), 255})
		}
	}
	d := &Display{}
	full := ComposeFullTable(field, d, [768]byte{}, 1, 2)
	if full.Rect.Size() != image.Pt(320, 609) {
		t.Fatal(full.Rect)
	}
	for y := 0; y < 576; y++ {
		if !bytes.Equal(field.Pix[y*field.Stride:(y+1)*field.Stride], full.Pix[(y+33)*full.Stride:(y+34)*full.Stride]) {
			t.Fatal("row identity", y)
		}
	}
}

func TestFullTableSourceNudgeKeepsMatrixFixed(t *testing.T) {
	field := image.NewRGBA(image.Rect(0, 0, 320, 576))
	for y := 0; y < 576; y++ {
		for x := 0; x < 320; x++ {
			o := field.PixOffset(x, y)
			copy(field.Pix[o:o+4], []byte{byte(y), byte(y >> 8), byte(x), 255})
		}
	}
	d := &Display{}
	d.Dots[160+5] = true
	var p [768]byte
	p[6] = 255
	zero := ComposeFullTable(field, d, p, 1, 2)
	if !bytes.Equal(zero.Pix, ComposeFullTableOffset(field, d, p, 1, 2, 0).Pix) {
		t.Fatal("zero-offset layout changed")
	}
	// TILT0 clamps SCREENPOS to 2048; SHR 9 exposes 0..4 source rows.
	for offset := int16(1); offset <= 4; offset++ {
		frame := ComposeFullTableOffset(field, d, p, 1, 2, offset)
		if !bytes.Equal(frame.Pix[:33*frame.Stride], zero.Pix[:33*zero.Stride]) {
			t.Fatal("nudge moved matrix")
		}
		for y := 0; y < 576-int(offset); y++ {
			if !bytes.Equal(frame.Pix[(y+33)*frame.Stride:(y+34)*frame.Stride], field.Pix[(y+int(offset))*field.Stride:(y+int(offset)+1)*field.Stride]) {
				t.Fatal("nudge direction/magnitude differs from SETSCREENSTART", offset, y)
			}
		}
		for y := 609 - int(offset); y < 609; y++ {
			if got := frame.RGBAAt(10, y); got.R != 0 || got.G != 0 || got.B != 0 || got.A != 255 {
				t.Fatal("vacated table rows must be opaque black", got)
			}
		}
	}
}
