//go:build windows

package platform

import (
	"pinballfantasies/internal/gamepad"
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

func TestSonyHIDYieldsToXInput(t *testing.T) {
	xinput := false
	h := &hostWindow{gamepad: windowsGamepad{readState: func(slot int) (xinputState, bool) { return xinputState{}, xinput && slot == 0 }}}
	h.gamepad.hid.present = true
	h.gamepad.hid.state = gamepad.Snapshot{Buttons: 1 << gamepad.LeftShoulder}
	h.pollGamepad()
	if len(h.events) != 2 || h.events[0].kind != 12 || h.events[0].key != 1<<gamepad.LeftShoulder || h.events[1] != (hostEvent{kind: 13, key: int(gamepad.PlayStation)}) {
		t.Fatal("Sony connection", h.events)
	}
	h.events = nil
	h.gamepad.hid.state = gamepad.Snapshot{}
	h.gamepad.lastPoll = time.Time{}
	h.pollGamepad()
	if len(h.events) != 1 || h.events[0].kind != 9 || h.events[0].key != gamepad.LeftShoulder || h.events[0].mouseY != 0 {
		t.Fatal("Sony release", h.events)
	}
	// A virtual XInput pad (Steam Input) takes over instead of doubling input.
	h.events = nil
	xinput = true
	h.gamepad.lastPoll, h.gamepad.nextScan = time.Time{}, time.Time{}
	h.pollGamepad()
	if len(h.events) != 3 || h.events[0].kind != 11 || h.events[1].kind != 12 || h.events[2] != (hostEvent{kind: 13, key: int(gamepad.Xbox)}) {
		t.Fatal("XInput takeover", h.events)
	}
	// Unplugging the Sony controller while XInput is active is silent.
	h.events = nil
	h.hidDeviceChange(gidcRemoval, h.gamepad.hid.device)
	h.gamepad.lastPoll = time.Time{}
	h.pollGamepad()
	if len(h.events) != 0 {
		t.Fatal("hidden device removal", h.events)
	}
}
