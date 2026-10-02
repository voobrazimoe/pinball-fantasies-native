package presentation

import (
	"pinballfantasies/internal/tablelogic"
	"strconv"
	"strings"
)

// Attract renders ShowHighsTS at a supplied frontend sync. It has no clock;
// repeated renders at the same counter return the same pixels. This program is
// separate from the table's suspended gameplay command/timing state.
func (d *Display) Attract(tick int, names [4]string, scores [4]string) *Display {
	out := *d
	out.Clear()
	out.On = true
	out.flashSpeed = 0
	out.Content.Texts = make(map[string][]byte, len(d.Content.Texts)+4)
	for k, v := range d.Content.Texts {
		out.Content.Texts[k] = v
	}
	for _, c := range d.Content.Attract {
		if strings.HasPrefix(c.Op, "_PRINT") && strings.Contains(c.Args[0], "HI_SCORE_LIST") {
			for rank := 0; rank < 4; rank++ {
				if strings.Contains(c.Args[0], "*"+strconv.Itoa(rank)+"+12") {
					out.Content.Texts[c.Args[0]] = append([]byte(names[rank]), 0)
				}
			}
		}
	}
	number := func(label string) string {
		for rank := 0; rank < 4; rank++ {
			if strings.Contains(label, "*"+strconv.Itoa(rank)+")") {
				return strings.TrimLeft(scores[rank], "0")
			}
		}
		return "0"
	}
	duration := func(c Command) int {
		duration := 1
		switch c.Op {
		case "_ANIMATION":
			a := out.Content.Animations[c.Args[0]]
			var frame uint16
			loops, time := a.Header[1], uint16(1)
			duration = 0
			for {
				duration++
				if tablelogic.Animation(a.Header, a.Durations, &frame, &loops, &time) {
					break
				}
			}
		case "_WAIT":
			duration = c.Num(0) + 1
		case "_CLEAR1":
			duration = 5
		case "_CLEAR2":
			duration = 17
		case "_CLEAR3":
			duration = 81
		case "_CLEAR4":
			duration = 5
		case "_SCROLL":
			duration = (len(out.Content.Texts[c.Args[0]])-21)*4 + 1
		case "_RULLGARDIN_UPP":
			duration = 17 - c.Num(1)
		case "_RULLGARDIN_NED":
			duration = 14 + c.Num(1)
		}
		return duration
	}
	// The source loop ends cleared, with FLASHOFF retained. Replay one cycle
	// rather than all elapsed frontend syncs, including after hours of attract.
	cycle := 0
	for _, c := range out.Content.Attract {
		cycle += duration(c)
	}
	if cycle > 0 {
		tick %= cycle
	}
	pc, left := 0, 0
	var frame, loops, frameTime uint16
	var anim Animation
	for t := 0; t <= tick; t++ {
		out.Flash()
		if left == 0 {
			c := out.Content.Attract[pc]
			pc = (pc + 1) % len(out.Content.Attract)
			out.BeginCommand(c)
			left = duration(c)
			if c.Op == "_ANIMATION" {
				anim = out.Content.Animations[c.Args[0]]
				frame = 0
				loops = anim.Header[1]
				frameTime = 1
			}
		}
		out.Visit(frame, 2, number)
		if out.op == "_ANIMATION" {
			tablelogic.Animation(anim.Header, anim.Durations, &frame, &loops, &frameTime)
		}
		left--
	}
	return &out
}
