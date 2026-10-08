//go:build dmoimpl1 || demodev

package partyland

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

func (d *demoConnected) admitGameplay(b, a []byte) error {

	font := d.game.Display.Content.Fonts["5"]
	if !bytes.Equal(a[font:font+42*5], b[0x19db0+0x6810:0x19db0+0x6810+42*5]) {
		return fmt.Errorf("font5 correspondence")
	}
	for i, label := range []string{"PLAYERSTEXT", "BALLSTEXT"} {
		at := 0x19db0 + []int{0x2379, 0x1e8b}[i]
		end := bytes.IndexByte(b[at:at+32], 0)
		if end < 0 {
			return fmt.Errorf("bounded panel text")
		}
		d.game.Display.Content.Texts[label] = append([]byte(nil), b[at:at+end+1]...)
		d.panelTexts[i] = append([]byte(nil), d.game.Display.Content.Texts[label]...)
	}
	d.gameplayFont = font
	d.gameplayOperands = true
	return nil
}

func (d *demoConnected) admitHighScore(b, a []byte) error {

	var top Decimal
	copy(top[:], b[0x19db0+0x16:0x19db0+0x16+12])
	d.factoryTop = &top
	d.factoryTopSource = "verified-native-factory-volatile"
	d.highScoreOperands = true
	return nil
}

func (d *demoConnected) admitScoredDrain(b, a []byte) error {

	if err := d.loadResetOperands(b); err != nil {
		return err
	}
	if err := d.loadExpiry(b, a); err != nil {
		return err
	}
	d.drainFont = d.game.Display.Content.Fonts["13"]
	if !bytes.Equal(a[d.drainFont:d.drainFont+42*13], b[0x200b0:0x200b0+42*13]) {
		return fmt.Errorf("drain font")
	}
	d.drainText = append([]byte(nil), b[0x1bb7e:0x1bb7e+17]...)
	d.game.Display.Content.Texts["DEMO_DRAIN_TEXT"] = append([]byte(nil), d.drainText...)
	d.drainBallText = append([]byte(nil), b[0x1c135:0x1c135+7]...)
	d.game.Display.Content.Texts["DEMO_BALLSTEXT"] = append([]byte(nil), d.drainBallText...)
	d.scoredDrainOperands = true
	return nil
}

func (d *demoConnected) admitInfo(b, a []byte) error {

	// Own the content maps as well as bytes; structural mutation must never
	// leak back into the shared canonical content cache or another candidate.
	fonts := make(map[string]int)
	for k, v := range d.game.Display.Content.Fonts {
		fonts[k] = v
	}
	d.game.Display.Content.Fonts = fonts
	texts := make(map[string][]byte)
	for k, v := range d.game.Display.Content.Texts {
		texts[k] = append([]byte(nil), v...)
	}
	d.game.Display.Content.Texts = texts
	d.infoTexts = make(map[string][]byte)
	refs := map[string]int{
		"PLAY_TEXT":       9017,
		"JACK_TEXT":       8986,
		"BONUS_TEXT":      8997,
		"ALLTIME_TEXT":    8678,
		"HI_1":            14131,
		"DEMO_HI_SCORE_0": 22,
		"DEMO_HI_NAME_0":  34,
		"HI_2":            14139,
		"DEMO_HI_SCORE_1": 38,
		"DEMO_HI_NAME_1":  50,
		"HI_3":            14147,
		"DEMO_HI_SCORE_2": 54,
		"DEMO_HI_NAME_2":  66,
		"HI_4":            14155,
		"DEMO_HI_SCORE_3": 70,
		"DEMO_HI_NAME_3":  82,
		"PLAYERSTEXT":     9081,
		"DEMO_BALLSTEXT":  9093,
	}
	for label, ref := range refs {
		at := 0x19db0 + ref
		n := 0
		if len(label) > 14 && label[:14] == "DEMO_HI_SCORE_" {
			n = 12
		} else {
			end := bytes.IndexByte(b[at:at+32], 0)
			if end < 0 {
				return fmt.Errorf("bounded info text")
			}
			n = end + 1
		}
		d.infoTexts[label] = append([]byte(nil), b[at:at+n]...)
		d.game.Display.Content.Texts[label] = append([]byte(nil), d.infoTexts[label]...)
	}
	copy(d.infoFactory[:], b[0x19dc6:0x19dc6+64])
	d.infoFactorySeal = sha256.Sum256(d.infoFactory[:])
	d.infoPresentation = true
	return nil
}
