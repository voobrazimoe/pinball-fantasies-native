package oracle

import "testing"

func TestOracleIdentityRemainsStrict(t *testing.T) {
	if len(pins) != 11 {
		t.Fatal("expected eleven immutable oracle identities")
	}
	if Known("PINBALL.CFG") {
		t.Fatal("mutable settings must not have an oracle payload gate")
	}
	for name, pin := range pins {
		if len(pin.SHA256) != 64 || pin.Size <= 0 {
			t.Fatal("invalid oracle pin", name)
		}
		if Verify(name, make([]byte, pin.Size)) == nil {
			t.Fatal("non-original content accepted by strict oracle", name)
		}
	}
	if Verify("unknown.PRG", nil) == nil {
		t.Fatal("unknown oracle identity accepted")
	}
}
