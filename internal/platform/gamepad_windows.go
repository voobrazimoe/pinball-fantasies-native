//go:build windows

package platform

import (
	"pinballfantasies/internal/gamepad"
	"syscall"
	"time"
	"unsafe"
)

type xinputState struct {
	Packet                    uint32
	Buttons                   uint16
	LeftTrigger, RightTrigger uint8
	LX, LY, RX, RY            int16
}
type windowsGamepad struct {
	proc               *syscall.LazyProc
	readState          func(int) (xinputState, bool)
	initialized        bool
	active             int
	lastPoll, nextScan time.Time
	tracker            gamepad.Tracker
	// Sony HID controllers are used only while no XInput slot is active, so a
	// controller that Steam Input also exposes as XInput is not read twice.
	hid       hidGamepad
	hidActive bool
}

func (p *windowsGamepad) read(slot int) (xinputState, bool) {
	if p.readState != nil {
		return p.readState(slot)
	}
	var state xinputState
	if p.proc == nil {
		return state, false
	}
	result, _, _ := p.proc.Call(uintptr(slot), uintptr(unsafe.Pointer(&state)))
	return state, result == 0
}
func (h *hostWindow) pollGamepad() {
	p := &h.gamepad
	now := time.Now()
	if !p.initialized {
		p.initialized = true
		p.active = -1
		for _, name := range []string{"xinput1_4.dll", "xinput9_1_0.dll"} {
			if p.readState != nil {
				break
			}
			proc := syscall.NewLazyDLL(name).NewProc("XInputGetState")
			if proc.Find() == nil {
				p.proc = proc
				break
			}
		}
	}
	xinput := p.proc != nil || p.readState != nil
	if (!xinput && !p.hid.present && !p.hidActive) || now.Sub(p.lastPoll) < 4*time.Millisecond {
		return
	}
	p.lastPoll = now
	var events []gamepad.Event
	if p.active >= 0 {
		state, ok := p.read(p.active)
		if ok {
			events = p.tracker.Sample(gamepad.XInputSnapshot(state.Buttons, state.LeftTrigger, state.RightTrigger))
		} else {
			events = p.tracker.Disconnect()
			p.active = -1
			p.nextScan = time.Time{}
		}
	}
	if xinput && p.active < 0 && !now.Before(p.nextScan) {
		p.nextScan = now.Add(time.Second)
		for slot := 0; slot < 4; slot++ {
			state, ok := p.read(slot)
			if ok {
				if p.hidActive {
					p.hidActive = false
					events = append(events, p.tracker.Disconnect()...)
				}
				p.active = slot
				events = append(events, p.tracker.Sample(gamepad.XInputSnapshot(state.Buttons, state.LeftTrigger, state.RightTrigger))...)
				break
			}
		}
	}
	if p.active < 0 {
		switch {
		case p.hid.present:
			p.hidActive = true
			events = append(events, p.tracker.Sample(p.hid.state)...)
		case p.hidActive:
			p.hidActive = false
			events = append(events, p.tracker.Disconnect()...)
		}
	}
	family := gamepad.Xbox
	if p.hidActive {
		family = gamepad.PlayStation
	}
	for _, event := range events {
		kind := 9 + event.Kind
		if event.Kind == 2 {
			kind = 12
		}
		if event.Kind == 3 {
			kind = 11
		}
		h.events = append(h.events, hostEvent{kind: kind, key: event.A, mouseY: event.B})
		if event.Kind == 2 {
			h.events = append(h.events, hostEvent{kind: 13, key: int(family)})
		}
	}
}
