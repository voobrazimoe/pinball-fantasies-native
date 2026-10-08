//go:build dmoimpl1

package partyland

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"reflect"
	"testing"
)

// Private data stays in memory. This is bounded input validation, not a DMO0
// search or a decoder/profile. The linked PRINT13 FONT_ADR is DS:6300.
func prepareDemoParty(t *testing.T, d *demoConnected) {
	t.Helper()
	b := pinnedDemoBytes(t)
	demoOK(t, d.loadResetOperands(b))
	a, err := os.ReadFile("../../TABLE1.PRG")
	if err != nil {
		t.Fatal(err)
	}
	demoOK(t, d.loadPartyPresentation(b, a))
}

func pinnedDemoBytes(t *testing.T) []byte {
	t.Helper()
	root := os.Getenv("PF_10MIN_DEMO_DATA")
	if root == "" {
		t.Skip("pinned demo presentation NOT AVAILABLE")
	}
	cmd := exec.Command("python3", "-c", `import os,sys
from pathlib import Path
sys.path.insert(0,'../../tools')
from audit_10min_demo import FILES,digest,require
root=Path(os.environ['PF_10MIN_DEMO_DATA'])
for name,(size,identity) in FILES.items():
 b=(root/name).read_bytes()
 require(len(b)==size and digest(b)==identity,'private demo identity mismatch')
`)
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Fatal("pinned demo validation", err)
	}
	b, err := os.ReadFile(filepath.Join(root, "TABLE1.PRG"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func prepareDemoReset(t *testing.T, d *demoConnected) {
	t.Helper()
	demoOK(t, d.loadResetOperands(pinnedDemoBytes(t)))
}

type demoFlashState struct {
	Speed, Count       uint16
	Enabled, Phase, On bool
}

func demoFlash(d *presentation.Display) demoFlashState {
	v := reflect.ValueOf(d).Elem()
	return demoFlashState{uint16(v.FieldByName("flashSpeed").Uint()), uint16(v.FieldByName("flashCount").Uint()), v.FieldByName("flashing").Bool(), v.FieldByName("flashPhase").Bool(), d.On}
}

func TestDemoPartyProgramBudgetTrue(t *testing.T) {
	d := demoDraining(t)
	prepareDemoParty(t, d)
	d.game.Session.CurrentPlayer = 2
	for i := range d.game.Display.Dots {
		d.game.Display.Dots[i] = true
	}
	wantOps := []string{"_CLEAR4", "_CLEAR4", "_CLEAR4", "_CLEAR4", "_FLASHON", "_PARTYONN", "_PRINT13", "_PARTYON"}
	for i, op := range wantOps {
		demoOK(t, d.sync(physics.Inputs{}, true))
		m := d.game.matrix
		if m.op != op || !m.active || d.game.waitCounters["PARTY_ON_TASK1"] != uint16(i+1) {
			t.Fatalf("calculation %d cursor/wait", i+1)
		}
		if i < 4 && (m.remaining != uint16(4-i) || m.next != 1) {
			t.Fatal("CLEAR4 visits", m)
		}
		if i == 4 && (m.next != 2 || demoFlash(d.game.Display) != (demoFlashState{3, 3, true, true, true}) || d.game.partyFlash) {
			t.Fatal("FLASHON dispatch")
		}
		if i == 5 && (m.next != 3 || !d.game.partyFlash || d.game.Display.Content.Texts["PARTY_ON_TEXT"][18] != byte(2)+'7' || demoFlash(d.game.Display).Count != 2) {
			t.Fatal("PARTYONN source writes")
		}
		if i == 4 || i == 5 {
			if d.game.Display.Dots != [presentation.DotWidth * presentation.DotHeight]bool{} {
				t.Fatal("CLEAR4 incomplete or premature PRINT")
			}
		}
		if i == 6 {
			expected := newTestGame(t).Display
			expected.Clear()
			expected.Content.Texts["PARTY_ON_TEXT"] = append([]byte(nil), d.game.Display.Content.Texts["PARTY_ON_TEXT"]...)
			expected.BeginCommand(demoPartyProgram()[3])
			expected.FlushPrint(nil)
			if m.next != 4 || d.game.Display.Dots == [presentation.DotWidth * presentation.DotHeight]bool{} || d.game.Display.Dots != expected.Dots {
				t.Fatal("actual PRINTTASK dot memory/position336")
			}
		}
		if i == 7 && (m.next != 5 || m.remaining != 1 || !d.game.partyFlash || demoFlash(d.game.Display) != (demoFlashState{3, 3, true, false, false})) {
			t.Fatal("PARTYRUT dispatch/blink phase")
		}
	}
	cursor, dots := d.game.matrix, d.game.Display.Dots
	// No matrix budget: blink still progresses, PARTYRUT and dot memory persist.
	for i := 0; i < 3; i++ {
		demoOK(t, d.sync(physics.Inputs{}, false))
	}
	if !reflect.DeepEqual(cursor, d.game.matrix) || dots != d.game.Display.Dots || !d.game.Display.On {
		t.Fatal("budget suppressed blink or altered cursor/dots")
	}
	for d.game.waitCounters["PARTY_ON_TASK1"] < 30 {
		demoOK(t, d.sync(physics.Inputs{}, true))
	}
	if !reflect.DeepEqual(cursor, d.game.matrix) {
		t.Fatal("PARTYRUT decremented or consumed terminator")
	}
	demoOK(t, d.sync(physics.Inputs{}, true))
	if d.handoffs != 1 || !d.game.partyFlash || !reflect.DeepEqual(cursor, d.game.matrix) || d.game.waitCounters["PARTY_ON_TASK1"] != 0 || d.game.Physics.Ball.PixelX != 282 {
		t.Fatal("flagged task handoff")
	}

}

func TestDemoPartyExpiryReplacement(t *testing.T) {
	for _, stage := range []int{5, 6, 7, 8} {
		t.Run(wantStage(stage), func(t *testing.T) {
			d := demoDraining(t)
			prepareDemoParty(t, d)
			d.counter = uint16(35997 - stage)
			for i := 0; i < stage; i++ {
				demoOK(t, d.sync(physics.Inputs{}, true))
			}
			old, dots, flash := d.game.matrix, d.game.Display.Dots, demoFlash(d.game.Display)
			party, text, slots := d.game.partyFlash, append([]byte(nil), d.game.Display.Content.Texts["PARTY_ON_TEXT"]...), d.game.taskIDs
			demoOK(t, d.sync(physics.Inputs{}, false))
			f := demoFlash(d.game.Display)
			if d.counter != 35998 || !d.expired || d.game.matrix.op != "_CLEAR4" || d.game.matrix.next != 1 || d.game.matrix.remaining != 5 || d.game.matrix.next == old.next || f.Enabled || !f.On || f.Count != flash.Count || f.Phase != flash.Phase || d.game.partyFlash != party || !bytes.Equal(text, d.game.Display.Content.Texts["PARTY_ON_TEXT"]) || d.game.taskIDs != slots || d.game.waitCounters["PARTY_ON_TASK1"] != uint16(stage+1) || d.game.Display.Dots != dots {
				t.Fatal("expiry replacement state")
			}
			for d.game.matrix.remaining > 1 {
				demoOK(t, d.sync(physics.Inputs{}, true))
			}
			before := d.game.matrix
			err := d.sync(physics.Inputs{}, true)
			if err == nil || d.failure.Producer != "NEXT_A" || !reflect.DeepEqual(before, d.game.matrix) || d.game.partyFlash != party {
				t.Fatal("expiry unsupported SCROLL/resumed old PARTY cursor")
			}
			assertConnectedSticky(t, d, err)
		})
	}
}
func wantStage(n int) string { return []string{"FLASHON", "PARTYONN", "PRINT13", "PARTYON"}[n-5] }

func TestDemoPartyBudgetFalseIndependentTask(t *testing.T) {
	d := demoDraining(t)
	prepareDemoParty(t, d)
	for i := 0; i < 30; i++ {
		demoOK(t, d.sync(physics.Inputs{}, false))
	}
	if d.game.waitCounters["PARTY_ON_TASK1"] != 30 || d.game.partyFlash || demoFlash(d.game.Display).Enabled || d.game.matrix.op != "_CLEAR4" || d.game.matrix.next != 1 || d.game.matrix.remaining != 5 {
		t.Fatal("budgetless matrix progressed")
	}
	before := d.game.matrix
	demoOK(t, d.sync(physics.Inputs{}, false))
	if d.handoffs != 1 || !d.game.partyFlash || d.game.waitCounters["PARTY_ON_TASK1"] != 0 || !reflect.DeepEqual(before, d.game.matrix) {
		t.Fatal("body store/reset/matrix preservation")
	}

}

func TestDemoPartyMissingPresentationBoundary(t *testing.T) {
	d := newDemoCore()
	demoOK(t, d.install(demoPartyProgram()))
	for i := 0; i < 5; i++ {
		demoOK(t, d.electronics(true))
	}
	if d.game.matrix.op != "_FLASHON" || d.game.matrix.next != 2 {
		t.Fatal("last CLEAR4 visit rejected early")
	}
	before := d.game.matrix
	if d.electronics(true) == nil || !reflect.DeepEqual(before, d.game.matrix) || d.game.partyFlash {
		t.Fatal("unverified PARTYONN consumed")
	}
}

func TestDemoPartyBudgetPauseAtEachConsumer(t *testing.T) {
	for _, stage := range []int{1, 5, 6, 7, 8} {
		d := demoDraining(t)
		prepareDemoParty(t, d)
		for i := 0; i < stage; i++ {
			demoOK(t, d.sync(physics.Inputs{}, true))
		}
		cursor, dots, party := d.game.matrix, d.game.Display.Dots, d.game.partyFlash
		text := append([]byte(nil), d.game.Display.Content.Texts["PARTY_ON_TEXT"]...)
		for i := 0; i < 3; i++ {
			demoOK(t, d.sync(physics.Inputs{}, false))
		}
		if !reflect.DeepEqual(cursor, d.game.matrix) || dots != d.game.Display.Dots || party != d.game.partyFlash || !bytes.Equal(text, d.game.Display.Content.Texts["PARTY_ON_TEXT"]) {
			t.Fatal("budgetless consumer mutation", stage)
		}
	}
}

func TestDemoPartyFalliblePrintAndTerminator(t *testing.T) {
	d := demoDraining(t)
	prepareDemoParty(t, d)
	for i := 0; i < 6; i++ {
		demoOK(t, d.sync(physics.Inputs{}, true))
	}
	// Structural malformed next operand after PARTYONN stores, before its WAIT visit.
	d.game.Display.Content.Commands[3].Nums[1] = 335
	before, dots, text := d.game.matrix, d.game.Display.Dots, append([]byte(nil), d.game.Display.Content.Texts["PARTY_ON_TEXT"]...)
	err := d.sync(physics.Inputs{}, true)
	if err == nil || d.failure.Producer != "NEXT_A" || !d.game.partyFlash || !reflect.DeepEqual(before, d.game.matrix) || dots != d.game.Display.Dots || !bytes.Equal(text, d.game.Display.Content.Texts["PARTY_ON_TEXT"]) {
		t.Fatal("PRINT gate lost preceding effects")
	}
	assertConnectedSticky(t, d, err)
	// Separate generic terminator contract: PARTYRUT itself never reaches zero.
	c := newDemoCore()
	demoOK(t, c.install([]presentation.Command{{Op: "_FLASHON", Args: []string{"3"}, Nums: map[int]int{0: 3}}, {Op: "0"}}))
	demoOK(t, c.electronics(true))
	f := demoFlash(c.game.Display)
	if c.game.matrix.active || c.game.matrix.op != "0" || c.game.matrix.next != 2 || f.Enabled || !f.On || f.Count != 2 {
		t.Fatal("ts_slut/terminator behavior")
	}
	beforeIdle := c.game.matrix
	demoOK(t, c.electronics(true))
	if !reflect.DeepEqual(beforeIdle, c.game.matrix) {
		t.Fatal("idle entered canonical fallback")
	}
}

func TestDemoPartyPresentationExtentRejects(t *testing.T) {
	d := newDemoCore()
	err := d.loadPartyPresentation(nil, nil)
	if err == nil || d.partyPresentation || d.game.matrix.active || d.game.partyFlash || d.loadPartyPresentation(nil, nil) != err {
		t.Fatal("missing assets admitted or rejection not sticky")
	}
}
