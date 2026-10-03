package platform

import "testing"

func TestWin32AltGrControlPair(t *testing.T) {
	ctrl := uintptr(0x1d << 16)
	alt := uintptr(0x38<<16 | 1<<24)
	for _, id := range []uint32{0x100, 0x101, 0x104, 0x105} {
		if !windowsAltGrControl(0x11, ctrl, 123, id, 0x12, alt, 123) {
			t.Fatalf("AltGr pair not recognized: message=%#x", id)
		}
	}
	for _, tc := range []struct {
		name       string
		vk         uint32
		l          uintptr
		id, nextVK uint32
		nextL      uintptr
		time       uint32
	}{
		{"independent Ctrl", 0x11, ctrl, 0x104, 0x12, alt, 124},
		{"right Ctrl", 0x11, ctrl | 1<<24, 0x104, 0x12, alt, 123},
		{"sided right Ctrl", 0xa3, ctrl | 1<<24, 0x104, 0x12, alt, 123},
		{"left Alt", 0x11, ctrl, 0x104, 0x12, 0x38 << 16, 123},
		{"other key", 0x11, ctrl, 0x100, 'A', alt, 123},
		{"focus message", 0x11, ctrl, 0x8, 0x12, alt, 123},
		{"other current key", 0x10, ctrl, 0x104, 0x12, alt, 123},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if windowsAltGrControl(tc.vk, tc.l, 123, tc.id, tc.nextVK, tc.nextL, tc.time) {
				t.Fatal("real input incorrectly suppressed")
			}
		})
	}
}

func TestWin32AltGrRepeatedGameplay(t *testing.T) {
	k := windowsKeys{}
	ctrl := uintptr(0x1d << 16)
	alt := uintptr(0x38<<16 | 1<<24)
	// Model ten minutes at 100 input cycles/second, including Ctrl/Alt polling
	// aliases, repeat messages, lost breaks, and periodic focus loss. Elapsed
	// wall time cannot affect this state machine; varying message times can.
	for cycle := uint32(0); cycle < 60000; cycle++ {
		stamp := cycle * 10
		if !windowsAltGrControl(0x11, ctrl, stamp, 0x104, 0x12, alt, stamp) {
			k.key(0x11, ctrl, true)
		}
		k.key(0x12, alt, true)
		for repeat := 0; repeat < 3; repeat++ {
			k.key(0x12, alt|1<<30, true)
			if got := k.reconcile(true, 1<<2|1<<4|1<<5); got != 2 {
				t.Fatalf("cycle=%d repeat=%d: Right Alt raised both flippers: %d", cycle, repeat, got)
			}
		}
		if cycle%7 != 0 { // Sometimes omit the Alt break entirely.
			k.key(0x12, alt|1<<30|1<<31, false)
		}
		if got := k.reconcile(cycle%101 != 0, 0); got != 0 {
			t.Fatalf("cycle=%d: release/focus loss left held input %d", cycle, got)
		}
	}
}

func TestWin32AltGrPreservesIndependentLeftFlipper(t *testing.T) {
	alt := uintptr(0x38<<16 | 1<<24)
	for _, vk := range []uint32{0xa0, 0xa2, 0xa4} {
		k := windowsKeys{}
		k.key(vk, 0, true)
		k.key(0x12, alt, true)
		if got := k.reconcile(true, 255); got != 3 {
			t.Fatalf("independent left modifier %#x lost: %d", vk, got)
		}
		k.key(0x12, alt, false)
		if got := k.reconcile(true, 255); got != 1 {
			t.Fatalf("Right Alt release lost real left modifier %#x: %d", vk, got)
		}
	}
}
