package presentation

import (
	"strconv"
	"strings"
)

// CheatTimeline owns the original demo-mode DO_MATRIX request. On completion,
// NODOT/SHOW_HI_ETC installs ShowHighsTS on the next sync. Both programs retain
// SCROLLE's phase and VGA memory; this clock advances only with frontend ticks.
type CheatTimeline struct {
	d   *Display
	run replay
}

func (d *Display) StartCheat(program string, names, scores [4]string) *CheatTimeline {
	for _, c := range d.Content.Attract {
		for rank := 0; rank < 4; rank++ {
			if strings.HasPrefix(c.Op, "_PRINT") && strings.Contains(c.Arg(0), "HI_SCORE_LIST") && strings.Contains(c.Arg(0), "*"+strconv.Itoa(rank)+"+12") {
				d.Content.Texts[c.Arg(0)] = append([]byte(names[rank]), 0)
			}
		}
	}
	t := &CheatTimeline{d: d}
	t.run = replay{d: d, commands: d.Content.Commands, active: true, loopHighs: true, pc: d.Content.Labels[program], numberFunc: func(label string) string {
		for rank := 0; rank < 4; rank++ {
			if strings.Contains(label, "HI_SCORE_LIST") && strings.Contains(label, "*"+strconv.Itoa(rank)+")") {
				return scores[rank]
			}
		}
		return "0"
	}}
	d.StartMatrix()
	t.run.dispatch()
	return t
}

func (t *CheatTimeline) Display() *Display { return t.d }
func (t *CheatTimeline) Tick() {
	if !t.run.active || t.run.handoff {
		t.d.StartMatrix()
		t.run.active, t.run.handoff = true, false
		t.run.pc = t.d.Content.Labels["SHOWHIGHSTS"]
		t.run.dispatch()
		return
	}
	t.run.step()
}
