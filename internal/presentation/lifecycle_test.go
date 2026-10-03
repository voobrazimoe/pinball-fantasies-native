package presentation

import (
	"fmt"
	"testing"
)

func TestGameOverContinuationUsesDemoNodotAndRetainsMemory(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			d := original(t, table)
			for i := range d.Dots {
				d.Dots[i] = i%3 == 0
			}
			d.scrollPhase = 3
			dots := d.Dots
			timeline := d.ContinueGameOver([]string{"12345"}, [4]string{}, [4]string{})
			if timeline.Display().Dots != dots || timeline.Display().scrollPhase != 3 {
				t.Fatal("handoff cleared retained memory")
			}
			timeline.Tick()
			if timeline.Display().CurrentOperation() != "_WAIT" || timeline.r.left != 20 {
				t.Fatal("demo NODOT did not dispatch MATRIXLGT then WAIT20", timeline.r.left)
			}
			if timeline.Display().Dots != dots {
				t.Fatal("initial handoff erased memory")
			}
			for visit := 1; visit < 20; visit++ {
				timeline.Tick()
				if timeline.r.left != 20-visit {
					t.Fatal("source wait cadence", visit)
				}
			}
			timeline.Tick()
			if timeline.Display().CurrentOperation() != "_SETLOOP" {
				t.Fatal("late WAIT20")
			}
			found := false
			for visit := 0; visit < 20; visit++ {
				timeline.Tick()
				if timeline.Display().CurrentOperation() == "_PRINT13" {
					if timeline.r.d.args[0] != "GAME_OVER_TEXT" {
						t.Fatal("wrong prompt")
					}
					found = true
					break
				}
			}
			if !found {
				t.Fatal("GAME OVER did not appear promptly")
			}
		})
	}
}

func TestContinuationRetainsZeroWaitWord(t *testing.T) {
	d := original(t, 4)
	d.Content.Commands = []Command{{Op: "0"}}
	d.BeginCommand(Command{Op: "_WAIT", Args: []string{"0"}, Nums: map[int]int{0: 0}})
	timeline := &Timeline{r: replay{d: d, commands: d.Content.Commands, left: WordWaitTicks(0), active: true}}
	for visit := 1; visit < 65536; visit++ {
		timeline.Tick()
		if !timeline.r.active {
			t.Fatalf("zero SI ended at visit %d", visit)
		}
	}
	timeline.Tick()
	if timeline.r.active {
		t.Fatal("zero SI did not finish on visit 65536")
	}
}
