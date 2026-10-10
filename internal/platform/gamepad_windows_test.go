//go:build windows

package platform

import (
	"testing"
	"time"
)

func TestXInputAlreadyConnectedAtStartup(t *testing.T) {
	state := xinputState{}
	calls := 0
	h := &hostWindow{events: []hostEvent{{kind: 8}}, gamepad: windowsGamepad{readState: func(slot int) (xinputState, bool) { calls++; return state, slot == 2 }}}
	if got := h.Event(); got.kind != 8 {
		t.Fatal(got)
	}
	if calls != 3 || h.gamepad.active != 2 {
		t.Fatal("startup enumeration", calls, h.gamepad.active)
	}
	h.events = focusLostEvents(h.events)
	for _, kind := range []int{5, 12, 13} {
		if got := h.Event(); got.kind != kind {
			t.Fatal("startup device lost across focus change", got)
		}
	}
	state.Buttons = 0x1000
	h.gamepad.lastPoll = time.Time{}
	if got := h.Event(); got.kind != 9 || got.key != 0 || got.mouseY != 1 {
		t.Fatal("first press needs reconnect", got)
	}
}
