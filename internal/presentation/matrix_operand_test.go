package presentation

import (
	"fmt"
	"os"
	"os/exec"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"strings"
	"testing"
)

// rendererNumericOperands lists the operand indices the shared presentation VM
// consumes as integers, with the original handler that reads them. It mirrors
// tools/matrix_operand_schema.py for the opcodes this package executes.
var rendererNumericOperands = map[string][]int{
	"_PRINT5":              {1},
	"_PRINT8":              {1},
	"_PRINT11":             {1},
	"_PRINT13":             {1},
	"_PRINT5_NUMBER":       {1},
	"_PRINT8_NUMBER":       {1},
	"_PRINT8_NUMBER_CENT":  {1},
	"_PRINT13_NUMBER":      {1},
	"_PRINT13_NUMBER_CENT": {1},
	"_RULLGARDIN_UPP":      {1},
	"_RULLGARDIN_NED":      {1},
	"_FLASHON":             {0},
	"_WAIT":                {0},
	"_TOWER":               {0},
}

func prg(t *testing.T, n int) []byte {
	t.Helper()
	testinputs.Require(t, fmt.Sprintf("../../TABLE%d.PRG", n))
	b, err := os.ReadFile(fmt.Sprintf("../../TABLE%d.PRG", n))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestMatrixOperandAuditFresh runs the source-authoritative operand checker
// during the normal suite. It recomputes every reachable operand from the
// original DOS ASM and fails on stale or unresolved generated operand data.
func TestMatrixOperandAuditFresh(t *testing.T) {
	testinputs.Require(t, "../../reference/original-dos-source/PLAND.ASM", "../../reference/original-dos-source/SDEV.ASM", "../../reference/original-dos-source/SHOW.ASM", "../../reference/original-dos-source/STONES.ASM", "../../TABLE1.PRG", "../../TABLE2.PRG", "../../TABLE3.PRG", "../../TABLE4.PRG")
	cmd := exec.Command("python3", "tools/matrix_operand_audit.py", "--check")
	cmd.Dir = "../.."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("source operand invariant: %v\n%s", err, out)
	}
}

// TestExtractedNumericOperandsResolved asserts the generated command stream
// carries a typed integer for every numeric operand the renderer consumes, and
// that the value agrees with the extractor's source provenance map. A stale
// content.go that predates the extraction fix fails here.
func TestExtractedNumericOperandsResolved(t *testing.T) {
	for n := 1; n <= 4; n++ {
		d := New(n, nil)
		sets := [][]Command{d.Content.Commands, d.Content.Attract}
		for si, set := range sets {
			for i, c := range set {
				for _, k := range rendererNumericOperands[c.Op] {
					if _, ok := c.Nums[k]; !ok {
						t.Errorf("table %d set %d command %d: %s operand %d (%q) has no resolved value",
							n, si, i, c.Op, k, c.Arg(k))
						continue
					}
					if got, want, ok := sourcePosition(d, c, k); ok && got != want {
						t.Errorf("table %d set %d command %d: %s operand %d resolved %d, source positions %q=%d",
							n, si, i, c.Op, k, got, c.Arg(k), want)
					}
				}
			}
		}
	}
}

// sourcePosition compares a resolved position operand against the generated
// provenance map, which holds the same source expression's assembler value.
// A disagreement means the extractor resolved the expression with the wrong
// assembler constants (the historical SW=320/336 confusion).
func sourcePosition(d *Display, c Command, k int) (int, int, bool) {
	if k >= len(c.Args) {
		return 0, 0, false
	}
	switch {
	case strings.HasPrefix(c.Op, "_PRINT"), strings.HasPrefix(c.Op, "_RULLGARDIN"), c.Op == "_TOWER":
	default:
		return 0, 0, false
	}
	want, ok := d.Content.Positions[c.Args[k]]
	if !ok {
		return 0, 0, false
	}
	return c.Num(k), want, true
}

// TestFlashOnUsesResolvedSourceSymbol exercises the real renderer path for the
// symbolic _FLASHON operands. Before the extraction fix these reached
// strconv.Atoi, whose error was dropped, so the matrix flashed at speed zero.
func TestFlashOnUsesResolvedSourceSymbol(t *testing.T) {
	cases := []struct {
		table int
		arg   string
		want  int
	}{
		{1, "DRSPEED", 13},
		{1, "LMSPEED", 7},
		{1, "PARTY_ON_SPEED", 3},
		{1, "SHOOT_AGAIN_SPEED", 3},
		{2, "PARTY_ON_SPEED", 3},
		{3, "PARTY_ON_SPEED", 3},
		{4, "PARTY_ON_SPEED", 3},
	}
	for _, tc := range cases {
		d := New(tc.table, nil)
		found := false
		for _, c := range append(append([]Command{}, d.Content.Commands...), d.Content.Attract...) {
			if c.Op != "_FLASHON" || c.Arg(0) != tc.arg {
				continue
			}
			found = true
			if got := c.Num(0); got != tc.want {
				t.Errorf("table %d: _FLASHON %s resolved %d, want %d", tc.table, tc.arg, got, tc.want)
				continue
			}
			d.BeginCommand(c)
			d.Visit(0, 0, nil)
			if int(d.flashSpeed) != tc.want {
				t.Errorf("table %d: _FLASHON %s set flash speed %d, want %d",
					tc.table, tc.arg, d.flashSpeed, tc.want)
			}
			break
		}
		if !found {
			t.Errorf("table %d: source _FLASHON %s not present in extracted content", tc.table, tc.arg)
		}
	}
}

// TestUnresolvedOperandStillPanics proves the invariant is retained: an operand
// that reached the runtime without a resolved value is a hard failure, never a
// default of zero and never a silent parse.
func TestUnresolvedOperandStillPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("unresolved operand did not panic")
		}
	}()
	Command{Op: "_WAIT", Args: []string{"2*60"}}.Num(0)
}

// TestSourceTextBytesMatchAssemblerExpressions pins the DATA buffers whose
// declarations mix character literals with arithmetic. A tokenizer that splits
// on quotes turns `DB '7'+2,'7 LITES EXTRA BALL'` into 0x37,0x02 and the matrix
// prints "0 0 LITES EXTRA BALL" instead of "20 LITES EXTRA BALL".
func TestSourceTextBytesMatchAssemblerExpressions(t *testing.T) {
	cases := []struct {
		table int
		label string
		want  string
	}{
		{2, "XB_AT_20_TEXT", "20 LITES EXTRA BALL"},
		{2, "OR_AT_10_TEXT", " 10 LITES OFF ROAD"},
		{3, "SKILLTEXT", "  MONEY MANIA AT 6 "},
		{4, "JUMP_AT_TEXT2", "XXX LITES 5 MILLIONS"},
		{1, "ZEROQ", "0"},
		{2, "ZEROQ", "0"},
		{3, "ZEROQ", "0"},
		{4, "ZEROQ", "0"},
	}
	for _, tc := range cases {
		d := New(tc.table, prg(t, tc.table))
		if got := strings.TrimRight(d.SourceText(tc.label), "\x00"); got != tc.want {
			t.Errorf("table %d %s renders %q, want %q", tc.table, tc.label, got, tc.want)
		}
	}
}

// TestAttractReplayUsesResolvedOperands drives the real attract replay across
// more than one cycle, so every _WAIT and _RULLGARDIN operand is consumed.
func TestAttractReplayUsesResolvedOperands(t *testing.T) {
	for n := 1; n <= 4; n++ {
		d := New(n, prg(t, n))
		names := [4]string{"AAA", "BBB", "CCC", "DDD"}
		scores := [4]string{"1000", "2000", "3000", "4000"}
		for tick := 0; tick < 4096; tick += 97 {
			d.Attract(tick, names, scores)
		}
	}
}

func TestSourceDBRowsAndRepeatLengths(t *testing.T) {
	d := New(3, prg(t, 3))
	want := []byte{0, 0, 0, 243, 243, 243, 243, 2, 5, 0, 0, 0}
	if !reflect.DeepEqual(d.Content.Texts["SPINSCORES"], want) {
		t.Fatal("SPINscores full source row")
	}
	for table := 1; table <= 4; table++ {
		d := New(table, prg(t, table))
		if len(d.Content.Texts["UNDANSPR"]) != 256 || len(d.Content.Texts["HIDDA"]) != 1152 {
			t.Fatal("DUP expression lengths", table)
		}
	}
}
func TestDisplayMissingTypedLiteralPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("missing typed literal was reparsed")
		}
	}()
	d := New(4, nil)
	d.BeginCommand(Command{Op: "_FLASHON", Args: []string{"3"}})
	d.Visit(0, 0, nil)
}

func TestMatrixLightUsesTypedOperand(t *testing.T) {
	d := New(4, nil)
	d.BeginCommand(Command{Op: "_MATRIXLGT", Args: []string{"0"}, Nums: map[int]int{0: 1}})
	d.Visit(0, 0, nil)
	if !d.On {
		t.Fatal("matrix flag consumed raw source instead of typed value")
	}
}
