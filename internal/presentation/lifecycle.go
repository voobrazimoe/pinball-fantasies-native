package presentation

import (
	"strconv"
	"strings"
)

// Timeline continues a live DOTRUT at a source-sync boundary. Unlike a frame
// projection it retains VGA memory, scroll phase and mutable score text.
type Timeline struct{ r replay }

func (t *Timeline) Tick() {
	if !t.r.active {
		// DEMOMODE NODOT returns SI=0 after installing AFTERDEMOMODETS.
		t.r.d.Flash()
		t.r.d.StartMatrix()
		t.r.pc = t.r.d.Content.Labels["AFTERDEMOMODETS"]
		t.r.active = true
		t.r.dispatch()
		t.r.dispatch()
		t.r.d.FlushPrint(t.r.number)
		return
	}
	t.r.step()
}
func (t *Timeline) Display() *Display { return t.r.d }
func (d *Display) ContinueGameOver(players []string, names, scores [4]string) *Timeline {
	out := *d
	out.Content.Texts = make(map[string][]byte, len(d.Content.Texts))
	for k, v := range d.Content.Texts {
		out.Content.Texts[k] = v
	}
	for _, c := range out.Content.Attract {
		if strings.HasPrefix(c.Op, "_PRINT") && strings.Contains(c.Arg(0), "HI_SCORE_LIST") {
			for rank := 0; rank < 4; rank++ {
				if strings.Contains(c.Arg(0), "*"+strconv.Itoa(rank)+"+12") {
					out.Content.Texts[c.Arg(0)] = append([]byte(names[rank]), 0)
				}
			}
		}
	}
	// _2_DEMO_MODE's unpreserved BX points into TASKLIST after ADDTASK.
	// HU_ reaches DUMRET, leaving DOTRUT=NODOT; the stream's long WAIT
	// is not installed. The demo task runs before this NODOT visit.
	t := &Timeline{r: replay{d: &out, commands: out.Content.Commands, scores: players, loopHighs: true}}
	t.r.numberFunc = func(label string) string {
		for rank := 0; rank < 4; rank++ {
			if strings.Contains(label, "*"+strconv.Itoa(rank)+")") {
				return strings.TrimLeft(scores[rank], "0")
			}
		}
		if label == "SIFFRORNA" && t.r.player >= 0 && t.r.player < len(players) {
			return players[t.r.player]
		}
		return "0"
	}
	return t
}

// HAJJSKAR and STJAERNOR are printed over retained VGA memory at DI=336.
// The keyboard routine updates the source text; it never clears a bounding box.
func (d *Display) ScoreEntry(player int, entry [3]byte, stars bool) {
	label := "HAJJSKAR"
	if stars {
		label = "STJAERNOR"
	} else {
		b := d.MutableText(label, 21)
		b[13] = byte(player) + '7'
		copy(b[16:19], entry[:])
	}
	d.Text(d.SourceText(label), 0, 1, 13)
}
