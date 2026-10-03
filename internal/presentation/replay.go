package presentation

import "strings"

// replay is the source DOTRUT/SISA/NEXT_A discipline for frontend-owned
// timelines. Dispatch writes illumination and installs PRINTTASK immediately;
// the newly installed routine is first visited on the following source sync.
// It consumes a supplied sync count and never reads host time.
type replay struct {
	d                             *Display
	commands                      []Command
	pc, left                      int
	frame, loops, timer, textLeft uint16
	anim                          Animation
	looper, player, shower        int
	scores                        []string
	active                        bool
	numberFunc                    func(string) string
	handoff                       bool
	loopHighs                     bool
}

func (r *replay) number(label string) string {
	if r.numberFunc != nil {
		return r.numberFunc(label)
	}
	if label == "SIFFRORNA" && len(r.scores) > 0 && r.player >= 0 && r.player < len(r.scores) {
		return r.scores[r.player]
	}
	return "0"
}
func (r *replay) dispatch() {
	for branches := 0; branches < 100; branches++ {
		c := r.commands[r.pc]
		r.pc++
		switch c.Op {
		case "0":
			r.active = false
			return
		case "_JMP":
			if c.Arg(0) == "SHOWHIGHSTS" && !r.loopHighs {
				r.handoff = true
				return
			}
			r.pc = r.d.Content.Labels[c.Arg(0)]
			continue
		case "_LOOP_":
			r.looper--
			if r.looper != 0 {
				r.pc = r.d.Content.Labels[c.Arg(1)]
				continue
			}
			r.looper = c.Num(0)
		case "_SETLOOP":
			r.looper = c.Num(0)
		case "_INIT_SCORE":
			r.player = -1
			r.shower = len(r.scores) + 1
		case "_SHOW_SCORE":
			r.player++
			r.shower--
			if r.shower > 0 {
				r.d.SetPlayers(r.player+1, len(r.scores), 1)
				if r.shower != 1 {
					r.pc = r.d.Content.Labels[c.Arg(0)]
				}
			} else {
				c = Command{Op: "_WAIT", Args: []string{"1"}, Nums: map[int]int{0: 1}}
			}
		}
		r.d.BeginCommand(c)
		r.left = 1
		switch c.Op {
		case "_WAIT":
			r.left = WordWaitTicks(c.Num(0))
		case "_CLEAR2":
			r.left = 17
		case "_CLEAR3":
			r.left = 81
		case "_CLEAR4":
			r.left = 5
		case "_RULLGARDIN_UPP":
			r.left = 16 - c.Num(1)
		case "_RULLGARDIN_NED":
			r.left = 13 + c.Num(1)
		case "_ANIMATION":
			r.anim = r.d.Content.Animations[c.Arg(0)]
			r.frame = 0
			r.loops = r.anim.Header[1]
			r.timer = 1
		case "_SCROLL":
			r.textLeft = uint16(len(r.d.Content.Texts[c.Arg(0)]) - 21)
		}
		return
	}
	panic("frontend source branch cycle")
}
func (r *replay) step() {
	defer r.d.FlushPrint(r.number)
	if !r.active {
		return
	}
	r.d.Flash()
	done := false
	switch r.d.op {
	case "_ANIMATION":
		done = r.d.StepAnimation(r.anim, &r.frame, &r.loops, &r.timer)
	case "_SCROLL":
		done = r.d.StepScroll(&r.textLeft)
	default:
		if !strings.HasPrefix(r.d.op, "_PRINT") {
			r.d.Visit(0, 0, r.number)
		}
		r.left--
		done = r.left == 0
	}
	if done {
		r.d.FinishRoutine(r.commands[r.pc].Op != "0")
		r.dispatch()
	}
}
