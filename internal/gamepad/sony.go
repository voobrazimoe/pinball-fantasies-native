package gamepad

// Sony controllers are not XInput devices. Windows exposes them as plain HID
// gamepads, so the Win32 host reads their input reports directly through raw
// input instead of depending on Steam Input, DS4Windows or SDL.

const sonyVendor = 0x054c

// SonyController reports whether a HID vendor/product pair is a DualSense,
// DualSense Edge or DualShock 4 whose input reports SonySnapshot understands.
func SonyController(vendor, product uint16) bool {
	if vendor != sonyVendor {
		return false
	}
	switch product {
	case 0x0ce6, 0x0df2, // DualSense, DualSense Edge
		0x05c4, 0x09cc, 0x0ba0: // DualShock 4 v1, v2, USB wireless adaptor
		return true
	}
	return false
}

func dualSense(product uint16) bool { return product == 0x0ce6 || product == 0x0df2 }

// SonySnapshot decodes one HID input report (report ID first) into the
// standard button snapshot used by polling hosts.
//
// Layouts, as offsets of the first stick byte:
//   - DualSense USB (0x01) and Bluetooth full report (0x31): sticks at 1/2,
//     triggers, a counter, then three button bytes.
//   - DualShock 4 USB (0x01), DualShock 4 Bluetooth (0x11) and DualSense
//     Bluetooth basic report (0x01, before any full-report request): sticks at
//     1/3, three button bytes, then triggers.
func SonySnapshot(product uint16, report []byte) (Snapshot, bool) {
	if len(report) < 10 {
		return Snapshot{}, false
	}
	base, triggersFirst := 0, false
	switch report[0] {
	case 0x01:
		// DualSense over Bluetooth sends a 10-byte basic report until a host
		// asks for the full 0x31 report; over USB the report is 64 bytes.
		base, triggersFirst = 1, dualSense(product) && len(report) >= 32
	case 0x31:
		base, triggersFirst = 2, true
	case 0x11:
		base = 3
	default:
		return Snapshot{}, false
	}
	var buttons, left, right int
	if triggersFirst {
		if len(report) < base+10 {
			return Snapshot{}, false
		}
		left, right = int(report[base+4]), int(report[base+5])
		buttons = base + 7
	} else {
		if len(report) < base+9 {
			return Snapshot{}, false
		}
		buttons = base + 4
		left, right = int(report[base+7]), int(report[base+8])
	}
	return Snapshot{
		Buttons:      sonyButtons(report[buttons], report[buttons+1], report[buttons+2]),
		LeftTrigger:  left * 32767 / 255,
		RightTrigger: right * 32767 / 255,
	}, true
}

func sonyButtons(b0, b1, b2 byte) uint16 {
	var s uint16
	set := func(on bool, button int) {
		if on {
			s |= 1 << button
		}
	}
	// Face buttons by position: Cross is A, Circle B, Square X, Triangle Y.
	set(b0&0x20 != 0, A)
	set(b0&0x40 != 0, B)
	set(b0&0x10 != 0, X)
	set(b0&0x80 != 0, Y)
	set(b1&0x01 != 0, LeftShoulder)
	set(b1&0x02 != 0, RightShoulder)
	set(b1&0x10 != 0, Back)  // Create / Share
	set(b1&0x20 != 0, Start) // Options
	set(b1&0x40 != 0, LeftStick)
	set(b1&0x80 != 0, RightStick)
	set(b2&0x01 != 0, Guide) // PS
	// The D-pad is a hat: 0 north, clockwise in eighths, 8 released.
	if hat := b0 & 0x0f; hat < 8 {
		set(hat == 7 || hat <= 1, Up)
		set(hat >= 1 && hat <= 3, Right)
		set(hat >= 3 && hat <= 5, Down)
		set(hat >= 5 && hat <= 7, Left)
	}
	return s
}
