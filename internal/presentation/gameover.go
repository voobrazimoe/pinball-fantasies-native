package presentation

// GameOver follows AfterDemoModeTS -> UrbanOverTS -> Once_MoreTS -> ShowHighsTS
// at the frontend's supplied source sync. No table tasks or scoring advance.
func (d *Display) GameOver(tick int, score string, names, scores [4]string) *Display {
	return d.GameOverPlayers(tick, []string{score}, names, scores)
}
func (d *Display) GameOverPlayers(tick int, players []string, names, scores [4]string) *Display {
	out := *d
	out.Clear()
	out.KillFlash()
	out.Content.Texts = make(map[string][]byte, len(d.Content.Texts))
	for k, v := range d.Content.Texts {
		out.Content.Texts[k] = v
	}
	r := replay{d: &out, commands: out.Content.Commands, pc: out.Content.Labels["AFTERDEMOMODETS"], scores: players, active: true}
	r.dispatch()
	for sync := 1; sync <= tick; sync++ {
		r.step()
		if r.handoff {
			return d.Attract(tick-sync, names, scores)
		}
	}
	return &out
}
