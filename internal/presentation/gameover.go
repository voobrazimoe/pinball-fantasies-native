package presentation

import "pinballfantasies/internal/tablelogic"

// GameOver renders the single-player AfterDemoModeTS -> UrbanOverTS ->
// Once_MoreTS -> ShowHighsTS program. Counter belongs to the frontend; this
// replay cannot advance table tasks or scores. Branches expand to the source's
// two player-score presentations, with WAIT n lasting n+1 visits.
func (d *Display) GameOver(tick int, score string, names, scores [4]string) *Display {
	return d.GameOverPlayers(tick, []string{score}, names, scores)
}

// GameOverPlayers expands SHOW_SCORE's player loop twice, in player order.
func (d *Display) GameOverPlayers(tick int, players []string, names, scores [4]string) *Display {
	clear := d.Content.Commands[d.Content.Labels["URBANOVERTS"]].Op
	type visit struct {
		op     string
		args   []string
		ticks  int
		player int
	}
	plan := []visit{{"_MATRIXLGT", []string{"0"}, 1, 0}, {"_WAIT", nil, 21, 0}, {"_SETLOOP", nil, 1, 0}, {"_INIT_SCORE", nil, 1, 0}}
	for loop := 0; loop < 2; loop++ {
		for player := range players {
			if player > 0 {
				plan = append(plan, visit{"_WAIT", nil, 81, player})
			}
			plan = append(plan, visit{clear, nil, 5, player}, visit{"_FLASHON", []string{"1"}, 1, player},
				visit{"_PRINT13", []string{"GAME_OVER_TEXT", "SW*2*2/4"}, 1, player}, visit{"_WAIT", nil, 31, player},
				visit{"_FLASHOFF", nil, 1, player}, visit{clear, nil, 5, player}, visit{"_SHOW_SCORE", nil, 1, player})
		}
		plan = append(plan, visit{"_LOOP", nil, 1, len(players) - 1})
		if loop == 0 {
			plan = append(plan, visit{"_INIT_SCORE", nil, 1, 0})
		}
		plan = append(plan, visit{"_WAIT", nil, 81, len(players) - 1})
	}
	plan = append(plan, visit{clear, nil, 5, 0}, visit{"_WAIT", nil, 21, 0})
	total := 0
	for _, p := range plan {
		total += p.ticks
	}
	if tick >= total {
		return d.Attract(tick-total, names, scores)
	}
	out := *d
	out.Clear()
	out.On = true
	out.flashSpeed = 0
	out.Content.Texts = make(map[string][]byte, len(d.Content.Texts))
	for k, v := range d.Content.Texts {
		out.Content.Texts[k] = v
	}
	pc, left := 0, 0
	player := 0
	var frame, loops, frameTime uint16
	var anim Animation
	for t := 0; t <= tick; t++ {
		out.Flash()
		if left == 0 {
			p := plan[pc]
			player = p.player
			out.SetPlayers(player+1, len(players), 1)
			pc++
			left = p.ticks
			if p.op == "_ANIMATION" {
				anim = out.Content.Animations["_CLEAR"]
				frame = 0
				loops = anim.Header[1]
				frameTime = 1
				out.Begin(p.op, []string{"_CLEAR"})
			} else {
				out.Begin(p.op, p.args)
			}
		}
		out.Visit(frame, 0, func(string) string { return players[player] })
		if out.op == "_ANIMATION" {
			tablelogic.Animation(anim.Header, anim.Durations, &frame, &loops, &frameTime)
		}
		left--
	}
	return &out
}
