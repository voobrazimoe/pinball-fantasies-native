package audio

import (
	"encoding/binary"
	"testing"
)

func structuralModule() []byte {
	b := make([]byte, 1084+1024+4)
	copy(b[1080:], "M.K.")
	b[950] = 15
	// One sample, two words. Every order points at the same stored pattern.
	binary.BigEndian.PutUint16(b[42:], 2)
	b[45] = 64
	binary.BigEndian.PutUint16(b[48:], 1)
	b[1086] = 0x10 // row 0/channel 0 uses sample 1.
	return b
}
func TestModuleCompatibilityStructure(t *testing.T) {
	b := structuralModule()
	if _, err := DecodeMenu(b); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"header truncation", func(b []byte) []byte { return b[:950] }},
		{"channel marker", func(b []byte) []byte { copy(b[1080:], "8CHN"); return b }},
		{"order count", func(b []byte) []byte { b[950] = 129; return b }},
		{"order extent", func(b []byte) []byte { b[950] = 14; return b }},
		{"pattern reference", func(b []byte) []byte { b[952] = 128; return b }},
		{"pattern truncation", func(b []byte) []byte { return b[:1100] }},
		{"sample numbering", func(b []byte) []byte { b[1084] = 0x20; return b }},
		{"missing sample", func(b []byte) []byte { binary.BigEndian.PutUint16(b[42:], 0); return b }},
		{"sample truncation", func(b []byte) []byte { return b[:len(b)-3] }},
		{"loop extent", func(b []byte) []byte { binary.BigEndian.PutUint16(b[48:], 3); return b }},
		{"volume", func(b []byte) []byte { b[45] = 65; return b }},
		{"finetune", func(b []byte) []byte { b[44] = 16; return b }},
		{"unsupported BPM", func(b []byte) []byte { b[1086] = 0x1f; b[1087] = 125; return b }},
		{"jump order", func(b []byte) []byte { b[1086] = 0x1b; b[1087] = 15; return b }},
		{"break row", func(b []byte) []byte { b[1086] = 0x1d; b[1087] = 0x64; return b }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := DecodeMenu(tc.mutate(append([]byte(nil), b...))); err == nil {
				t.Fatal("incompatible MOD accepted")
			}
		})
	}
	b[0] ^= 1
	b[len(b)-1] ^= 1
	if _, err := DecodeMenu(b); err != nil {
		t.Fatal("compatible title/audio rejected", err)
	}
}
