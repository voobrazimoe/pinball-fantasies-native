//go:build dmoimpl1

package partyland

import (
	"fmt"
	"os"
	"os/exec"
	"pinballfantasies/internal/presentation"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func (d *demoConnected) loadInfoOperands(t *testing.T) {
	t.Helper()
	d.loadScoredDrainOperands(t)
	cmd := exec.Command("python3", "../../tools/check_demo_2b12_consumers.py")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ShowInfoTS admission: %v %s", err, out)
	}
	b := pinnedDemoBytes(t)
	a, err := os.ReadFile("../../TABLE1.PRG")
	if err != nil {
		t.Fatal(err)
	}
	demoOK(t, d.admitInfo(b, a))
}

func structuralInfo(t *testing.T) *demoConnected {
	t.Helper()
	d := connected(newTestGame(t))
	d.loadInfoOperands(t)
	d.infoCount = 720
	d.game.Physics.SpringValid = false
	d.game.inChute = false
	d.game.Score = Number(50030)
	d.game.SetHighScore(Number(0))
	d.game.Display.Content.Texts["BONUS_TEXT"][11] = '8'
	return d
}
func TestDemoInfoCompleteStructural(t *testing.T) {
	d := structuralInfo(t)
	g := d.game
	d.expired = true
	d.counter = 36725
	audio, waits, ids := g.Audio, fmt.Sprint(g.waitCounters), g.taskIDs
	demoOK(t, d.gameplayIdle())
	if d.infoStarts != 1 || d.infoCount != 1 || !d.fjantText || d.dotReady || g.matrix.next != 2 || g.matrix.op != "_CLEAR4" || g.matrix.remaining != 5 {
		t.Fatal("entry/same visit dispatch", g.matrix)
	}
	program := demoInfoProgram()
	seen := []string{}
	ticks := 0
	previous := -1
	painted := map[string]bool{}
	for g.matrix.active && ticks < 2000 {
		pc := g.matrix.next - 1
		if pc != previous {
			seen = append(seen, program[pc].Op)
			previous = pc
		}
		before := g.Display.Dots
		// Independently construct each pending PRINTTASK result from the linked
		// text/record and raw position. No expected frame is supplied to execution.
		expected := *g.Display
		checkPrint := false
		if g.matrix.remaining == 1 && g.matrix.next < len(program) {
			next := program[g.matrix.next]
			if strings.HasPrefix(next.Op, "_PRINT") {
				checkPrint = true
				raw := next.Nums[1]
				row := (raw + 84) / 168
				x, y := (raw-row*168)*2, row-1
				height := 13
				if next.Op == "_PRINT5" {
					height = 5
				}
				label := next.Args[0]
				if next.Op == "_PRINT13_NUMBER" {
					value := g.Jackpot.Uint64()
					for rank, v := range []uint64{50000000, 25000000, 10000000, 5000000} {
						if label == "DEMO_HI_SCORE_"+strconv.Itoa(rank) {
							value = v
						}
					}
					expected.Number(strconv.FormatUint(value, 10), x, y, height)
				} else {
					expected.Text(strings.TrimRight(g.Display.SourceText(label), "\x00"), x, y, height)
				}
			}
		}
		demoOK(t, d.taskMatrixSuffix(true))
		if checkPrint && expected.Dots != g.Display.Dots {
			t.Fatal("PRINTTASK dot memory/data mapping", g.matrix.next)
		}
		ticks++
		if before != g.Display.Dots {
			painted[program[pc].Op] = true
		}
		if !d.expired || d.counter != 36725 || d.infoCount != 1 || g.Score != Number(50030) || g.Bonus != Number(0) || g.Audio != audio || fmt.Sprint(g.waitCounters) != waits || g.taskIDs != ids {
			t.Fatal("unrelated state mutation")
		}
	}
	if g.matrix.active || g.matrix.next != len(program) || g.matrix.op != "0" || ticks != 1072 {
		t.Fatal("not complete", ticks, g.matrix)
	}
	want := []string{}
	for _, c := range program[1 : len(program)-1] {
		want = append(want, c.Op)
	}
	if !reflect.DeepEqual(seen, want) {
		t.Fatal("dispatch order", seen, want)
	}
	for _, op := range []string{"_CLEAR4", "_RULLGARDIN_NED", "_RULLGARDIN_UPP", "_PRINT13", "_PRINT13_NUMBER", "_PRINT5"} {
		if !painted[op] {
			t.Fatal("no real dot effects", op)
		}
	}
	if demoFlash(g.Display).Enabled {
		t.Fatal("terminal flash survived")
	}
	demoOK(t, d.gameplayIdle())
	if d.infoCount != 2 || g.matrix.next != 2 || g.matrix.op != "_PRINT5" {
		t.Fatal("return idle panel", g.matrix)
	}
	t.Logf("ShowInfoTS all %d commands; routine visits=%d; real dot effects; next idle panel", len(program), ticks)
}
func TestDemoInfoGuardsStructural(t *testing.T) {
	for _, mode := range []string{"719", "720", "721", "spring", "no-admission", "missing-text", "changed-text", "missing-name", "changed-name", "missing-score", "changed-score", "changed-record", "persistent", "font5", "font13", "bad-bcd", "active"} {
		t.Run(mode, func(t *testing.T) {
			d := structuralInfo(t)
			g := d.game
			switch mode {
			case "719":
				d.infoCount = 719
			case "721":
				d.infoCount = 721
			case "spring":
				g.Physics.SpringValid = true
				d.infoTexts = nil
				g.Display.Content.Texts["DEMO_HI_NAME_0"] = nil
			case "no-admission":
				d.infoPresentation = false
			case "missing-text":
				g.Display.Content.Texts["PLAY_TEXT"] = nil
			case "changed-text":
				g.Display.Content.Texts["JACK_TEXT"][0] ^= 1
			case "missing-name":
				g.Display.Content.Texts["DEMO_HI_NAME_3"] = nil
			case "changed-name":
				g.Display.Content.Texts["DEMO_HI_NAME_2"][0] ^= 1
			case "missing-score":
				g.Display.Content.Texts["DEMO_HI_SCORE_3"] = nil
			case "changed-score":
				g.Display.Content.Texts["DEMO_HI_SCORE_1"][0] = 9
			case "changed-record":
				d.infoFactory[63] ^= 1
			case "persistent":
				d.factoryTopSource = "persistent"
			case "font5":
				g.Display.Content.Fonts["5"]++
			case "font13":
				g.Display.Content.Fonts["13"]++
			case "bad-bcd":
				g.Jackpot[0] = 10
			case "active":
				g.matrix.active = true
			}
			before := fmt.Sprintf("%#v %#v %#v", g, g.Display, g.Physics)
			count := d.infoCount
			fjant := d.fjantText
			err := d.showInfoPreflight()
			valid := mode == "720" || mode == "spring"
			if (err == nil) != valid || before != fmt.Sprintf("%#v %#v %#v", g, g.Display, g.Physics) || count != d.infoCount || fjant != d.fjantText {
				t.Fatal("preflight", err)
			}
			if err != nil {
				assertConnectedSticky(t, d, err)
				if d.showInfo() != err {
					t.Fatal("direct sticky")
				}
				return
			}
			demoOK(t, d.gameplayIdle())
			if mode == "spring" && (d.infoStarts != 0 || d.infoCount != 721 || g.matrix.active) {
				t.Fatal("guard bypass")
			}
		})
	}
}
func TestDemoInfoBudgetPauseReplacementStructural(t *testing.T) {
	d := structuralInfo(t)
	g := d.game
	// A matrix budget miss cannot consume the NODOT threshold.
	g.matrixTimeLeft = false
	demoOK(t, d.matrixVisit())
	if d.infoCount != 720 || d.infoStarts != 0 {
		t.Fatal("budget idle")
	}
	g.matrixTimeLeft = true
	demoOK(t, d.gameplayIdle())
	for i := 0; i < 25; i++ {
		demoOK(t, d.matrixVisit())
	}
	matrix := g.matrix
	dots := g.Display.Dots
	flash := demoFlash(g.Display)
	count := d.infoCount
	g.matrixTimeLeft = false
	demoOK(t, d.matrixVisit())
	if !reflect.DeepEqual(matrix, g.matrix) || dots != g.Display.Dots || flash != demoFlash(g.Display) || count != d.infoCount {
		t.Fatal("budget active")
	}
	demoOK(t, d.calculation(true, true))
	if !reflect.DeepEqual(matrix, g.matrix) || dots != g.Display.Dots || flash != demoFlash(g.Display) || count != d.infoCount {
		t.Fatal("pause")
	}
	d.expired = true
	g.matrixTimeLeft = true
	demoOK(t, d.install(demoExpiryProgram()))
	if !d.expired || g.matrix.next != 1 || g.matrix.op != "_CLEAR4" {
		t.Fatal("expiry replacement")
	}
	// Replacement does not save the info cursor; usual guards still apply.
	demoOK(t, d.install(demoScoredPlayerProgram()))
	if !d.expired || g.matrix.next != 1 || g.matrix.op != "_CLEAR4" || demoFlash(g.Display).Enabled {
		t.Fatal("replacement")
	}
}
func TestDemoInfoNestedConsumerStructural(t *testing.T) {
	d := structuralInfo(t)
	demoOK(t, d.gameplayIdle())
	g := d.game
	g.Display.Content.Commands[g.matrix.next] = presentation.Command{Op: "_DOBEATEN"}
	for g.matrix.remaining > 1 {
		demoOK(t, d.matrixVisit())
	}
	before := g.Display.Dots
	matrix := g.matrix
	err := d.matrixVisit()
	if err == nil || before != g.Display.Dots || !reflect.DeepEqual(matrix, g.matrix) || g.ExtraBalls != 0 {
		t.Fatal("nested consumer effects", err)
	}
	assertConnectedSticky(t, d, err)
}

func TestDemoInfoBudgetAtEveryConsumerStructural(t *testing.T) {
	d := structuralInfo(t)
	g := d.game
	demoOK(t, d.gameplayIdle())
	previous := -1
	for n := 0; g.matrix.active && n < 2000; n++ {
		if g.matrix.next != previous {
			previous = g.matrix.next
			matrix, dots, flash := g.matrix, g.Display.Dots, demoFlash(g.Display)
			demoOK(t, d.calculation(true, true))
			if !reflect.DeepEqual(matrix, g.matrix) || dots != g.Display.Dots || flash != demoFlash(g.Display) {
				t.Fatal("pause", previous)
			}
			expected := *g.Display
			expected.Flash()
			demoOK(t, d.taskMatrixSuffix(false))
			if !reflect.DeepEqual(matrix, g.matrix) || dots != g.Display.Dots || demoFlash(&expected) != demoFlash(g.Display) {
				t.Fatal("matrix budget/independent flash", previous)
			}
		}
		demoOK(t, d.taskMatrixSuffix(true))
	}
	if g.matrix.active {
		t.Fatal("bounded info did not complete")
	}
}

func TestDemoInfoTerminalIndependentStructural(t *testing.T) {
	d := structuralInfo(t)
	demoOK(t, d.gameplayIdle())
	demoOK(t, d.install(demoExpiryProgram()))
	for d.terminal == nil && d.visits < 2000 {
		demoOK(t, d.taskMatrixSuffix(true))
	}
	if d.terminal == nil || d.terminal.Status != 0 {
		t.Fatal("expiry QUIT missing")
	}
	before := fmt.Sprintf("%#v %#v %#v", d, d.game, d.game.Display)
	demoOK(t, d.showInfo())
	demoOK(t, d.showInfoPreflight())
	demoOK(t, d.matrixVisit())
	if before != fmt.Sprintf("%#v %#v %#v", d, d.game, d.game.Display) {
		t.Fatal("info revived terminal")
	}
}
