package gamepad

import "testing"

func TestSonyControllerIdentities(t *testing.T) {
	for _, pid := range []uint16{0x0ce6, 0x0df2, 0x05c4, 0x09cc, 0x0ba0} {
		if !SonyController(0x054c, pid) {
			t.Fatalf("%#04x not recognised", pid)
		}
	}
	if SonyController(0x045e, 0x0ce6) || SonyController(0x054c, 0x0268) {
		t.Fatal("foreign or unsupported device accepted")
	}
}

func TestDualSenseUSBReport(t *testing.T) {
	r := make([]byte, 64)
	r[0] = 0x01
	r[5], r[6] = 255, 0 // L2, R2
	r[8] = 0x20 | 2     // Cross, D-pad east
	r[9] = 0x01 | 0x20  // L1, Options
	r[10] = 0x01        // PS
	s, ok := SonySnapshot(0x0ce6, r)
	want := uint16(1<<A | 1<<Right | 1<<LeftShoulder | 1<<Start | 1<<Guide)
	if !ok || s.Buttons != want || s.LeftTrigger != 32767 || s.RightTrigger != 0 {
		t.Fatalf("got %+v ok=%t, want buttons %#x", s, ok, want)
	}
}

func TestDualSenseBluetoothReports(t *testing.T) {
	full := make([]byte, 78)
	full[0] = 0x31
	full[7] = 128      // R2
	full[9] = 0x40 | 8 // Circle, D-pad released
	full[10] = 0x02    // R1
	s, ok := SonySnapshot(0x0ce6, full)
	if !ok || s.Buttons != 1<<B|1<<RightShoulder || s.RightTrigger != 128*32767/255 {
		t.Fatalf("full report: %+v", s)
	}
	basic := []byte{0x01, 128, 128, 128, 128, 0x10 | 6, 0x10, 0x00, 255, 0}
	s, ok = SonySnapshot(0x0ce6, basic)
	if !ok || s.Buttons != 1<<X|1<<Left|1<<Back || s.LeftTrigger != 32767 {
		t.Fatalf("basic report: %+v", s)
	}
}

func TestDualShock4Reports(t *testing.T) {
	usb := make([]byte, 64)
	usb[0] = 0x01
	usb[5] = 0x80 | 7 // Triangle, D-pad north-west
	usb[6] = 0x02     // R1
	usb[9] = 255      // R2
	s, ok := SonySnapshot(0x09cc, usb)
	if !ok || s.Buttons != 1<<Y|1<<Up|1<<Left|1<<RightShoulder || s.RightTrigger != 32767 {
		t.Fatalf("USB: %+v", s)
	}
	bt := make([]byte, 78)
	bt[0] = 0x11
	bt[7] = 0x20 | 4 // Cross, D-pad south
	s, ok = SonySnapshot(0x09cc, bt)
	if !ok || s.Buttons != 1<<A|1<<Down {
		t.Fatalf("Bluetooth: %+v", s)
	}
}

func TestSonyRejectsUnknownReports(t *testing.T) {
	for _, r := range [][]byte{nil, {0x01, 0, 0}, append([]byte{0x05}, make([]byte, 63)...)} {
		if _, ok := SonySnapshot(0x0ce6, r); ok {
			t.Fatalf("accepted %v", r)
		}
	}
}
