package platform

import "testing"

func TestWin32LostBreakReconciliation(t *testing.T) {
	for i, vk := range windowsHeldVKs {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			k := windowsKeys{}
			l := uintptr(0)
			if vk == 40 {
				l = 1 << 24
			}
			k.key(vk, l, true) // Deliberately never deliver the break.
			if k.held() == 0 {
				t.Fatal("make did not set history")
			}
			if got := k.reconcile(true, 0); got != 0 || k.held() != 0 || k.down[vk] {
				t.Fatalf("lost break stuck: physical held=%d history=%d", got, k.held())
			}
		})
	}
}

func TestWin32PhysicalModifierCombinations(t *testing.T) {
	// Exhaust every combination, including sided, simultaneous and controls
	// pressed without a make being received by this window.
	for mask := 0; mask < 256; mask++ {
		k := windowsKeys{}
		want := 0
		if mask&0x15 != 0 {
			want |= 1
		}
		if mask&0x2a != 0 {
			want |= 2
		}
		if mask&0x40 != 0 {
			want |= 4
		}
		if mask&0x80 != 0 {
			want |= 8
		}
		if got := k.reconcile(true, uint8(mask)); got != want {
			t.Fatalf("physical=%#x held=%d want=%d", mask, got, want)
		}
		if got := k.reconcile(false, uint8(mask)); got != 0 {
			t.Fatalf("global keys leaked without focus: %#x held=%d", mask, got)
		}
	}
	k := windowsKeys{}
	for _, vk := range windowsHeldVKs {
		k.key(vk, 1<<24, true)
	}
	// Releasing just left Shift must leave left Ctrl/Alt and right controls held.
	if got := k.reconcile(true, 0xfe); got != 15 || k.down[0xa0] || !k.down[0xa1] {
		t.Fatal("one modifier released other simultaneous modifiers")
	}
	k.key('A', 0, true)
	if got := k.reconcile(false, 255); got != 0 || k != (windowsKeys{}) {
		t.Fatal("focus loss must clear the entire message history")
	}
	if got := k.reconcile(true, 0); got != 0 {
		t.Fatal("focus regain resurrected stale input")
	}
}

func TestWin32ReconcileStaleAltBeforeEnter(t *testing.T) {
	for _, vk := range []uint32{0xa4, 0xa5} {
		k := windowsKeys{}
		k.key(vk, 0, true)
		k.reconcile(true, 0)
		e := k.key(13, 0x1c<<16, true)[0]
		keys := hostKeys{}
		if e.alt || keys.down(e.key, e.alt, e.repeat) != dispatchKey {
			t.Fatal("stale Alt turned plain Enter into Alt+Enter")
		}
		// A queued real WM_SYSKEYDOWN retains its message-time Alt context,
		// even if both keys have already been released by the time it is pumped.
		k.key(13, 0, false)
		e = k.key(13, 1<<29|0x1c<<16, true)[0]
		keys.enterUp()
		if keys.down(e.key, e.alt, e.repeat) != toggleFullscreen {
			t.Fatal("reconciliation lost legitimate Alt+Enter")
		}
	}
}

func TestWin32ReconcilePreservesEvents(t *testing.T) {
	k := windowsKeys{}
	for _, tc := range []struct {
		vk   uint32
		l    uintptr
		code int
	}{
		{32, 0, 57}, {40, 1 << 24, 80}, {13, 0, 28},
		{112, 0, 59}, {113, 0, 60}, {114, 0, 61}, {115, 0, 62}, {116, 0, 63}, {117, 0, 64}, {118, 0, 65}, {119, 0, 66}, {'Q', 16 << 16, 16},
	} {
		k.reconcile(true, 255) // physical polling must not synthesize makes/repeats
		e := k.key(tc.vk, tc.l, true)[0]
		if e.key != tc.code || e.repeat {
			t.Fatalf("first make changed: %+v", e)
		}
		e = k.key(tc.vk, tc.l|1<<30, true)[0]
		if !e.repeat {
			t.Fatal("repeat lost")
		}
		k.reconcile(true, 0)
		events := k.key(tc.vk, tc.l, false)
		if tc.vk == 13 || tc.vk == 40 {
			want := 4
			if tc.vk == 40 {
				want = 3
			}
			if len(events) != 1 || events[0].kind != want {
				t.Fatal("release event lost", events)
			}
		} else if len(events) != 0 {
			t.Fatal("extra release event", events)
		}
	}
}
