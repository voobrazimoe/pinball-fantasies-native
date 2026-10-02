//go:build windows

package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"pinballfantasies/internal/gameplay"
	"time"
)

var asyncKeyState = up("GetAsyncKeyState")
var inputFocus = up("GetFocus")
var inputForeground = up("GetForegroundWindow")

func (h *hostWindow) refreshInput(source string, vk uint32, l uintptr, makeKey bool) int {
	var physical uint8
	for i, key := range windowsHeldVKs {
		state, _, _ := asyncKeyState.Call(uintptr(key))
		// Only the high bit reports current state. The low bit is unreliable
		// "pressed since last call" history, not a held control.
		if state&0x8000 != 0 {
			physical |= 1 << i
		}
	}
	focus, _, _ := inputFocus.Call()
	foreground, _, _ := inputForeground.Call()
	focused := h.focused && h.hwnd != 0 && focus == h.hwnd && foreground == h.hwnd
	before := h.input.held()
	held := h.input.reconcile(focused, physical)
	if h.inputLog != nil {
		edge := "-"
		if source == "key" {
			edge = "break"
			if makeKey {
				edge = "make"
			}
		}
		fmt.Fprintf(h.inputLog, "%s source=%s vk=%#02x scan=%#02x extended=%t edge=%s alt-context=%t previous=%t focused=%t physical=%#02x held=%#x message-held=%#x\n",
			time.Now().Format(time.RFC3339Nano), source, vk, (l>>16)&255,
			l&(1<<24) != 0, edge, l&(1<<29) != 0, l&(1<<30) != 0, focused, physical, held, before)
	}
	return held
}

func (h *hostWindow) openInputLog() {
	if os.Getenv("PF12_INPUT_LOG") != "1" {
		return
	}
	path := os.Getenv("PF12_INPUT_LOG_PATH")
	if path == "" {
		var dir string
		var err error
		if diagnosticFile != nil {
			dir = filepath.Dir(diagnosticFile.Name())
		} else {
			dir, err = StateDirectory("")
			if err == nil {
				err = os.MkdirAll(dir, 0700)
			}
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "PF12 input trace:", err)
			return
		}
		path = filepath.Join(dir, "input.log")
	}
	var err error
	h.inputLog, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		fmt.Fprintln(os.Stderr, "PF12 input trace:", err)
		return
	}
	fmt.Fprintln(h.inputLog, "PF12 physical bits: 0=LShift 1=RShift 2=LCtrl 3=RCtrl 4=LAlt 5=RAlt 6=Down 7=Space; held: 1=Left 2=Right 4=Down 8=Space")
}

// Controls translates host held state into logical source controls.
func (h *hostWindow) Controls() gameplay.Controls {
	bits := h.Held()
	return gameplay.Controls{Left: bits&1 != 0, Right: bits&2 != 0, Down: bits&4 != 0, Tilt: bits&8 != 0}
}
