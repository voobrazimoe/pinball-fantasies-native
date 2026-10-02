package presentation

import (
	"os/exec"
	"pinballfantasies/internal/testinputs"
	"testing"
)

// TestSourceReachableMatrixRegistered is the in-tree half of the
// source-authority check. tools/matrix_reachability.py derives the reachable
// program set from the original DOS ASM; this asserts the generated content
// registers every one of them. It is the check that would have caught the
// missing SHOW Beaten_bh_TS program.
func TestSourceReachableMatrixRegistered(t *testing.T) {
	for n, want := range sourceReachableMatrix {
		d := New(n, nil)
		for _, label := range want {
			if _, ok := d.Content.Labels[label]; !ok {
				t.Errorf("table %d: source-reachable matrix program %s is not registered", n, label)
			}
		}
	}
}

// TestMatrixProgramTargetsResolve proves the registered program stream is
// closed: every branch target, animation and scroll it names exists.
func TestMatrixProgramTargetsResolve(t *testing.T) {
	for n := 1; n <= 4; n++ {
		d := New(n, nil)
		for i, c := range d.Content.Commands {
			a := c.Args
			arg := func(k int) string {
				if k < len(a) {
					return a[k]
				}
				return ""
			}
			switch c.Op {
			case "_BEATEN_MATRIX":
				if _, ok := d.Content.Labels["BEATEN_BH_TS"]; !ok {
					t.Errorf("table %d: beaten implicit target absent", n)
				}
			case "_CHECK_XXBALLS", "_KOLLA_XXBALL":
				targets := []string{"SHOOT_AGAIN_ONTS", "CHECK_XXBALLTS", "AFTER_XXBALLTS"}
				if n == 4 {
					targets[2] = "AFTER_XXBALLSTS"
				}
				for _, label := range targets {
					if _, ok := d.Content.Labels[label]; !ok {
						t.Errorf("table %d: match implicit target %s absent", n, label)
					}
				}
			case "_TOWER":
				if _, ok := d.Content.Positions[arg(0)]; !ok {
					t.Errorf("Tower operand %s unresolved", arg(0))
				}
			case "_JMP", "_JBONUSX1", "_SHOW_SCORE":
				if _, ok := d.Content.Labels[arg(0)]; !ok {
					t.Errorf("table %d command %d: %s branches to unregistered %q", n, i, c.Op, arg(0))
				}
			case "_JBCDZ", "_LOOP_":
				if _, ok := d.Content.Labels[arg(1)]; !ok {
					t.Errorf("table %d command %d: %s branches to unregistered %q", n, i, c.Op, arg(1))
				}
			case "_ANIMATION":
				if _, ok := d.Content.Animations[arg(0)]; !ok {
					t.Errorf("table %d command %d: unknown animation %q", n, i, arg(0))
				}
			case "_SCROLL":
				if _, ok := d.Content.Texts[arg(0)]; !ok {
					t.Errorf("table %d command %d: unknown scroll text %q", n, i, arg(0))
				}
			}
		}
	}
}

// Recompute from the original ASM during the normal suite so a stale generated
// reachable set cannot accidentally hide a source/extractor regression.
func TestSourceMatrixAuditFresh(t *testing.T) {
	testinputs.Require(t, "../../reference/original-dos-source/PLAND.ASM", "../../reference/original-dos-source/SDEV.ASM", "../../reference/original-dos-source/SHOW.ASM", "../../reference/original-dos-source/STONES.ASM", "../../TABLE1.PRG", "../../TABLE2.PRG", "../../TABLE3.PRG", "../../TABLE4.PRG")
	cmd := exec.Command("python3", "tools/matrix_reachability.py", "--check")
	cmd.Dir = "../.."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("source matrix invariant: %v\n%s", err, out)
	}
}
