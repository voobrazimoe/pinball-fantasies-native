package presentation

import (
	"pinballfantasies/internal/gameplay"
	"testing"
)

func TestCommonCheatProgramsUseSourceScrollLifecycle(t *testing.T) {
	// FANTASIE TECHTS..FAIRPLAYTS: each table's CLEARIT, _SCROLL, 0.
	// SCROLLE calls twice per sync, DEC byte phase, SI+20 boundary test.
	for table := 1; table <= 4; table++ {
		for _, cheat := range gameplay.OriginalCheats {
			d := original(t, table)
			pos, ok := d.Content.Labels[cheat.Program]
			if !ok {
				t.Fatal("missing original common program", table, cheat.Program)
			}
			clear, scroll, end := d.Content.Commands[pos], d.Content.Commands[pos+1], d.Content.Commands[pos+2]
			if scroll.Op != "_SCROLL" || end.Op != "0" {
				t.Fatal("common program extraction")
			}
			raw := d.Content.Texts[scroll.Arg(0)]
			if len(raw) < 43 || raw[len(raw)-1] != 255 {
				t.Fatal("common linked scroll terminator")
			}
			for i := 0; i < 21; i++ {
				if raw[i] != 1 || raw[len(raw)-2-i] != 1 {
					t.Fatal("common scroll uses encoded byte1 blanks")
				}
			}
			duration := 1
			if clear.Op == "_CLEAR4" {
				duration = 5
			}
			if clear.Op == "_ANIMATION" {
				a := d.Content.Animations[clear.Arg(0)]
				// _CLEAR headers/frame durations come from the source DATA2.
				duration = 1
				for _, ticks := range a.Durations {
					duration += int(ticks) * int(a.Header[1])
				}
			}
			timeline := d.StartCheat(cheat.Program, [4]string{}, [4]string{})
			for tick := 1; tick <= duration; tick++ {
				timeline.Tick()
			}
			if d.op != "_SCROLL" || !timeline.run.active {
				t.Fatal("CLEARIT completion", table, cheat.Program, duration, d.op)
			}
			calls := 4*(len(raw)-21) + 1
			for tick := 1; tick < calls; tick++ {
				timeline.Tick()
				if !timeline.run.active || d.op != "_SCROLL" {
					t.Fatal("early common scroll termination", table, cheat.Program, tick)
				}
			}
			timeline.Tick()
			if timeline.run.active {
				t.Fatal("scroll completion boundary", table, cheat.Program, calls)
			}
			retained, phase := d.Dots, d.scrollPhase
			timeline.Tick()
			if !timeline.run.active || timeline.run.pc != d.Content.Labels["SHOWHIGHSTS"]+1 || d.Dots != retained || d.scrollPhase != phase {
				t.Fatal("NODOT installs highs next sync retaining VGA/SCROLLE", table, cheat.Program)
			}
		}
	}
}
