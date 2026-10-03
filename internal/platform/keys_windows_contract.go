package platform

// Win32 message decoding is kept pure for make/break and layout tests on Linux.
// Virtual letters match SDL's physical US scan contract (DOS initials/control codes).
type windowsKeys struct{ down [256]bool }

// Windows implements AltGr as Left Ctrl followed by extended Right Alt with
// the same message timestamp. Identify the paired Ctrl before it enters the
// flipper history; suppressing all Ctrl while Right Alt is held would also
// suppress a real, independently pressed left flipper.
func windowsAltGrControl(vk uint32, l uintptr, stamp uint32, nextID, nextVK uint32, nextL uintptr, nextStamp uint32) bool {
	if (vk != 0x11 && vk != 0xa2) || l&(1<<24) != 0 {
		return false
	}
	switch nextID {
	case 0x100, 0x101, 0x104, 0x105:
		return (nextVK == 0x12 || nextVK == 0xa5) && nextL&(1<<24) != 0 && stamp == nextStamp
	}
	return false
}

// Query the sided VKs, never VK_SHIFT/VK_CONTROL/VK_MENU. Each index in the
// snapshot has its own bit, so overlapping modifiers can be reconciled
// independently.
var windowsHeldVKs = [...]uint32{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 40, 32}

// reconcile repairs lost breaks in the message history, then returns the
// message-authoritative gameplay controls. Physical polling is deliberately
// release-only: GetAsyncKeyState must never invent a make for the opposite side
// (or activate a control before its queued make message is decoded).
func (k *windowsKeys) reconcile(focused bool, physical uint8) int {
	if !focused {
		*k = windowsKeys{}
		return 0
	}
	for i, vk := range windowsHeldVKs {
		if physical&(1<<i) == 0 {
			k.down[vk] = false
		}
	}
	// Only the sided identities are stored by key().
	k.down[0x10], k.down[0x11], k.down[0x12] = false, false, false
	return k.held()
}

func (k *windowsKeys) key(vk uint32, l uintptr, makeKey bool) []hostEvent {
	if vk == 0x10 {
		vk = 0xa0
		if (l>>16)&255 == 0x36 {
			vk = 0xa1
		}
	}
	if vk == 0x11 {
		vk = 0xa2
		if l&(1<<24) != 0 {
			vk = 0xa3
		}
	}
	if vk == 0x12 {
		vk = 0xa4
		if l&(1<<24) != 0 {
			vk = 0xa5
		}
	}
	// Keep keypad Enter/navigation distinct, matching SDL physical scancodes.
	if vk == 13 && l&(1<<24) != 0 {
		vk = 0xfe
	}
	if vk == 38 && l&(1<<24) == 0 {
		vk = 0xfd
	}
	if vk == 40 && l&(1<<24) == 0 {
		vk = 0xfc
	}
	if vk >= 256 {
		return nil
	}
	repeat := k.down[vk] || l&(1<<30) != 0
	k.down[vk] = makeKey
	if !makeKey {
		if vk == 13 {
			return []hostEvent{{kind: 4}}
		}
		if vk == 40 {
			return []hostEvent{{kind: 3}}
		}
		return nil
	}
	code := 127
	switch vk {
	case 27:
		code = 1
	case 13:
		code = 28
	case 32:
		code = 57
	case 38:
		code = 72
	case 40:
		code = 80
	case 106:
		code = 55
	case 112, 113, 114, 115, 116, 117, 118, 119:
		code = 59 + int(vk-112)
	}
	letters := "QWERTYUIOPASDFGHJKLZXCVBNM"
	scans := [...]int{16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 30, 31, 32, 33, 34, 35, 36, 37, 38, 44, 45, 46, 47, 48, 49, 50}
	for i := range letters {
		if vk == uint32(letters[i]) {
			code = scans[i]
		}
	}
	// SDL consumes physical letter scancodes rather than translated characters.
	if scan := int((l >> 16) & 255); vk >= 65 && vk <= 90 && scan != 0 {
		for _, s := range scans {
			if scan == s {
				code = scan
				break
			}
		}
	}
	return []hostEvent{{kind: 2, key: code, alt: k.down[0xa4] || k.down[0xa5] || l&(1<<29) != 0, repeat: repeat}}
}
func (k *windowsKeys) held() int {
	bits := 0
	if k.down[0xa0] || k.down[0xa2] || k.down[0xa4] {
		bits |= 1
	}
	if k.down[0xa1] || k.down[0xa3] || k.down[0xa5] {
		bits |= 2
	}
	if k.down[40] {
		bits |= 4
	}
	if k.down[32] {
		bits |= 8
	}
	return bits
}
