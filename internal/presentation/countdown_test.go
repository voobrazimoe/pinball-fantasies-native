package presentation

import (
	"fmt"
	"testing"
)

func TestCountdownSourcePrintSlotAndUninstall(t *testing.T) {
	// FANTASIE _COUNTDOWN initializes SEC_ASC to operand+1; COUNTDOWN
	// decrements on visits 1,72,..., shows zero, then uninstalls 71 later.
	for table := 1; table <= 4; table++ {
		for _, seconds := range []int{0, 2, 10, 25, 30, 50} {
			t.Run(fmt.Sprintf("table%d-%d", table, seconds), func(t *testing.T) {
				d := original(t, table)
				for i := range d.Dots {
					d.Dots[i] = true
				}
				d.StartCountdown(seconds/10, seconds%10)
				want := *d
				number := func(string) string { return "1234567" }
				calls := 0
				for tick := 1; tick <= seconds*71+72; tick++ {
					previous := d.Dots
					done := d.StepCountdown("BCD", false, func(sec int) {
						calls++
						if sec != seconds-(tick-1)/71 {
							t.Fatal("READ_SPECIAL_MODE_COUNTER digits", tick, sec)
						}
					})
					if d.Dots != previous {
						t.Fatal("COUNTDOWN drew before PRINTTASK", tick)
					}
					if done != (tick == seconds*71+72) {
						t.Fatal("uninstall boundary", tick, done)
					}
					if done {
						if d.pendingPrint != nil {
							t.Fatal("uninstall installed print")
						}
						break
					}
					if (tick-1)%71 == 0 {
						sec := seconds - (tick-1)/71
						want.Text(fmt.Sprintf("%2d", sec), 144, 2, 11)
					} else {
						want.Number("1234567", 16, 1, 13)
					}
					d.FlushPrint(number)
					if d.Dots != want.Dots {
						t.Fatal("source exclusive print slot pixels", tick)
					}
				}
				if calls != seconds+1 {
					t.Fatal("unexpected callback count", calls)
				}
			})
		}
	}
}

func TestCountdownContinuationInhibitionAndReplacement(t *testing.T) {
	for table := 1; table <= 4; table++ {
		d := original(t, table)
		d.StartCountdown(2, 5)
		d.StepCountdown("BCD", false, nil)
		d.FlushPrint(func(string) string { return "1" })
		for i := 0; i < 17; i++ {
			d.StepCountdown("BCD", false, nil)
			d.FlushPrint(func(string) string { return "1" })
		}
		state := d.countdown
		for i := 0; i < 100; i++ {
			d.StepCountdown("BCD", true, nil)
			d.FlushPrint(func(string) string { return "2" })
		}
		if d.countdown != state {
			t.Fatal("INH_CD advanced seconds or phase")
		}
		d.BeginCommand(Command{Op: "_WAIT", Nums: map[int]int{0: 1}})
		if d.countdown != state {
			t.Fatal("replacement discarded SEC_ASC/SYNC_LEFT")
		}
		d.ContinueCountdown()
		d.StepCountdown("BCD", false, func(sec int) {
			if sec != 24 {
				t.Fatal("continuation did not decrement retained seconds", sec)
			}
			// READ_SPECIAL_MODE_COUNTER can accept a new effect. The following
			// seconds PRINTTASK wins over a print installed by that effect.
			d.BeginCommand(Command{Op: "_PRINT5", Args: []string{"PLAYERSTEXT", "340"}, Nums: map[int]int{1: 340}})
		})
		if d.pendingPrint.Op != "_PRINT11" || d.pendingPrint.Arg(0) != "SEC_ASC" {
			t.Fatal("callback replaced final seconds task")
		}
	}
}
