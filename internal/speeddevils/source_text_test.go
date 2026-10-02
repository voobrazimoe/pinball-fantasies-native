package speeddevils

import "testing"

// TestMilesChainTextsMatchSource covers the two defects reported on the left
// (Miles) ramp:
//
//  1. SDEV.ASM:1819-1820 declare the threshold texts with assembler byte
//     expressions, `DB '7'+2,'7 LITES EXTRA BALL'` and
//     `DB ' ','7'+1,'7 LITES OFF ROAD'`, whose encoded digits were being split
//     into two bytes by the extractor, so the matrix printed
//     "0 0 LITES EXTRA BALL" / "0 0 LITES OFF ROAD".
//  2. SDEV.ASM:4773-4774 (NotStandardSeries -> Put_In_Text) write the live mile
//     counter into `miles_text`, replacing its "XXX"; the port printed the raw
//     template.
func TestMilesChainTextsMatchSource(t *testing.T) {
	g := testGame(t)
	quiet(g)
	for _, tc := range []struct {
		label string
		want  string
	}{
		{"XB_AT_20_TEXT", "20 LITES EXTRA BALL\x00"},
		{"OR_AT_10_TEXT", " 10 LITES OFF ROAD\x00"},
		{"MILES_TEXT", "     XXX MILES      \x00"},
		{"JUMP_AT_TEXT", "XXX LITES THE JUMP\x00"},
		{"OFFROAD_AT_TEXT", "XXX LITES OFF ROAD\x00"},
	} {
		if got := g.Display.SourceText(tc.label); got != tc.want {
			t.Fatalf("%s: source bytes render %q, want %q", tc.label, got, tc.want)
		}
	}
}

// TestMilesCountersWrittenIntoSourcetexts drives the real award path and checks
// the digits the original writes into the shared text buffers. '*' is the
// original's blank for a leading zero; the font has no glyph for it.
func TestMilesCountersWrittenIntoSourcetexts(t *testing.T) {
	for _, tc := range []struct {
		miles    uint16
		nextJump uint16
		nextOff  uint16
		milesTxt string
		jumpTxt  string
		offTxt   string
	}{
		{0, 20, 20, "     *** MILES      \x00", "*20 LITES THE JUMP\x00", "*20 LITES OFF ROAD\x00"},
		{7, 20, 15, "     **7 MILES      \x00", "*20 LITES THE JUMP\x00", "*15 LITES OFF ROAD\x00"},
		{20, 40, 20, "     *20 MILES      \x00", "*40 LITES THE JUMP\x00", "*20 LITES OFF ROAD\x00"},
		{123, 140, 120, "     123 MILES      \x00", "140 LITES THE JUMP\x00", "120 LITES OFF ROAD\x00"},
	} {
		g := testGame(t)
		quiet(g)
		g.Miles, g.NextJump, g.NextOffRoad = tc.miles, tc.nextJump, tc.nextOff
		g.writeMilesText()
		g.writeJumpText()
		g.writeOffRoadText()
		if got := g.Display.SourceText("MILES_TEXT"); got != tc.milesTxt {
			t.Errorf("miles=%d: MILES_TEXT %q, want %q", tc.miles, got, tc.milesTxt)
		}
		if got := g.Display.SourceText("JUMP_AT_TEXT"); got != tc.jumpTxt {
			t.Errorf("jump=%d: JUMP_AT_TEXT %q, want %q", tc.nextJump, got, tc.jumpTxt)
		}
		if got := g.Display.SourceText("OFFROAD_AT_TEXT"); got != tc.offTxt {
			t.Errorf("offroad=%d: OFFROAD_AT_TEXT %q, want %q", tc.nextOff, got, tc.offTxt)
		}
	}
}

// TestAwardPathWritesMilesText runs the real loop-award routine so the
// write-back is exercised from gameplay rather than only through the helper.
func TestAwardPathWritesMilesText(t *testing.T) {
	g := testGame(t)
	quiet(g)
	g.Miles = 19
	g.NextJump, g.NextOffRoad = 40, 30
	g.loop(true)
	if g.Miles != 20 {
		t.Fatalf("miles %d, want 20", g.Miles)
	}
	if got := g.Display.SourceText("MILES_TEXT"); got != "     *20 MILES      \x00" {
		t.Fatalf("MILES_TEXT after award %q", got)
	}
}

// TestTextWriteBackDoesNotLeakIntoOtherGames guards the shared extracted byte
// slices: a second game must still see the untouched source template.
func TestTextWriteBackDoesNotLeakIntoOtherGames(t *testing.T) {
	a := testGame(t)
	quiet(a)
	a.Miles = 42
	a.writeMilesText()
	if got := a.Display.SourceText("MILES_TEXT"); got != "     *42 MILES      \x00" {
		t.Fatalf("first game MILES_TEXT %q", got)
	}
	b := testGame(t)
	quiet(b)
	if got := b.Display.SourceText("MILES_TEXT"); got != "     XXX MILES      \x00" {
		t.Fatalf("template leaked between games: %q", got)
	}
}
