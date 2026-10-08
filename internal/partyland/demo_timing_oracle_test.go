//go:build dmoimpl1

package partyland

import (
	"bytes"
	"fmt"
	"testing"
)

// Independent linked SCROLLE oracle: literal SI+20 FF test before EACH
// subcall, shared phase decrement after glyph stores, SI increment on zero.
// Does not call StepScroll, ScrollCompletes, or use matrix.textLeft for control.
type demoTimingOracle struct {
	phase, si, pc, left, visits int
	texts                       [2][]byte
	done                        bool
}

func (o *demoTimingOracle) dispatch() {
	o.pc++
	c := demoExpiryProgram()[o.pc-1]
	switch c.Op {
	case "_CLEAR4":
		o.left = 5
	case "_SCROLL":
		o.si = 0
	case "_FLASHON", "_FLASHOFF", "_PRINT13_NUMBER":
		o.left = 1
	case "_WAIT", "_DEMO_FADE":
		o.left = c.Num(0)
	case "_DEMO_QUIT":
		o.done = true
	default:
		panic(c.Op)
	}
}
func (o *demoTimingOracle) restart() { o.pc = 0; o.done = false; o.visits = 0; o.dispatch() }
func (o *demoTimingOracle) visit() {
	o.visits++
	c := demoExpiryProgram()[o.pc-1]
	if c.Op == "_SCROLL" {
		i := 0
		if o.pc == 7 {
			i = 1
		}
		for sub := 0; sub < 2; sub++ {
			if o.texts[i][o.si+20] == 255 {
				o.dispatch()
				return
			}
			o.phase--
			if o.phase == 0 {
				o.phase = 8
				o.si++
			}
		}
	} else {
		o.left--
		if o.left == 0 {
			o.dispatch()
		}
	}
}
func TestDemoTimingIndependentFFOracle(t *testing.T) {
	for _, restart := range []bool{false, true} {
		t.Run(fmt.Sprint(restart), func(t *testing.T) {
			d := expiryFixture(t)
			o := demoTimingOracle{phase: 8, texts: d.expiryTexts}
			for i, text := range o.texts {
				if len(text) != []int{99, 89}[i] || bytes.IndexByte(text, 255) != len(text)-1 {
					t.Fatal("private FF extent")
				}
			}
			o.restart()
			previous := 0
			for !o.done {
				if restart && d.counter == 36172 {
					// Real queued replacement, before this calculation's matrix admission.
					demoOK(t, d.queue(demoTask{Site: "timing-oracle-restart", Action: demoReplaceMatrix, Program: demoExpiryProgram()}))
					t.Logf("inherited phase=%d SI=%d", o.phase, o.si)
					o.restart()
					previous = 0
				}
				o.visit()
				demoOK(t, d.calculation(false, true))
				m := d.game.matrix
				if m.next != o.pc {
					t.Fatalf("calculation %d command %d oracle %d", d.counter, m.next, o.pc)
				}
				if m.op == "_SCROLL" {
					i := 0
					if o.pc == 7 {
						i = 1
					}
					want := bytes.IndexByte(o.texts[i], 255) - o.si - 20
					if int(m.textLeft) != want {
						t.Fatalf("SI+20 remaining at %d: %d want %d", d.counter, m.textLeft, want)
					}
				} else if !o.done && int(m.remaining) != o.left {
					t.Fatal("remaining", d.counter, m.remaining, o.left)
				}
				if previous != o.pc {
					t.Logf("boundary calculation=%d command=%d op=%s visits=%d phase=%d", d.counter, o.pc, m.op, o.visits, o.phase)
					previous = o.pc
				}
				if o.visits > 1200 {
					t.Fatal("oracle did not terminate")
				}
			}
			wantVisits, wantQuit := 1050, 37047
			if restart {
				wantVisits, wantQuit = 1048, 37220
			}
			if d.terminal == nil || o.visits != wantVisits || int(d.counter) != wantQuit {
				t.Fatal("terminal", d.terminal, o.visits)
			}
		})
	}
}
