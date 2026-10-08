//go:build dmoimpl1 || demodev

package partyland

// Unpacked BCD, unsigned most-significant-first. Equality does not qualify.
func demoScoreAbove(score, top Decimal) (bool, bool) {
	for i := range score {
		if score[i] > 9 || top[i] > 9 {
			return false, false
		}
	}
	for i := range score {
		if score[i] != top[i] {
			return score[i] > top[i], true
		}
	}
	return false, true
}

func (d *demoCore) checkHighScorePreflight() error {
	if d.failure != nil {
		return d.failure
	}
	g := d.game
	if !d.highScoreOperands {
		return d.reject("NODOT", "CHECKHIGHSCORE operands", "not admitted")
	}

	if g.Score != Number(0) && d.infoCount == 720 && !d.infoPresentation {
		return d.reject("NODOT", "SHOW_HI_ETC", "inactivity high-score presentation not admitted")
	}

	if g.inChute || d.specialMode || g.alreadyBeaten {
		return nil
	}
	if d.factoryTopSource != "verified-native-factory-volatile" || d.factoryTop == nil {
		return d.reject("NODOT", "CHECKHIGHSCORE source", "missing or persistent/unverified top-score state")
	}
	qualifies, valid := demoScoreAbove(g.Score, *d.factoryTop)
	if !valid || *d.factoryTop != Number(50000000) {
		return d.reject("NODOT", "CHECKHIGHSCORE source", "invalid or changed factory score")
	}
	if qualifies {

		return d.reject("CHECKHIGHSCORE", "BEATENTS/_DOBEATEN", "reward consumer not admitted")
	}
	return nil
}
