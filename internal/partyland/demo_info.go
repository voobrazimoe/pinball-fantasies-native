//go:build dmoimpl1 || demodev

package partyland

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"pinballfantasies/internal/presentation"
	"reflect"
	"strconv"
)

// Complete linked ShowInfoTS, including the expanded historical showithi macro.
// This is command metadata only; strings, names and glyphs are owner-local.
func demoInfoProgram() []presentation.Command {
	return []presentation.Command{
		{Op: "_CLEAR4"},
		{Op: "_CLEAR4"},
		{Op: "_RULLGARDIN_NED", Args: []string{"PLAY_TEXT", "1"}, Nums: map[int]int{1: 1}},
		{Op: "_WAIT", Args: []string{"120"}, Nums: map[int]int{0: 120}},
		{Op: "_MATRIXLGT", Args: []string{"0"}, Nums: map[int]int{0: 0}},
		{Op: "_CLEAR4"},
		{Op: "_PRINT13", Args: []string{"JACK_TEXT", "336"}, Nums: map[int]int{1: 336}},
		{Op: "_PRINT13_NUMBER", Args: []string{"JACKVALUE", "368"}, Nums: map[int]int{1: 368}},
		{Op: "_FLASHON", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_WAIT", Args: []string{"120"}, Nums: map[int]int{0: 120}},
		{Op: "_FLASHOFF", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_CLEAR4"},
		{Op: "_RULLGARDIN_UPP", Args: []string{"BONUS_TEXT", "1"}, Nums: map[int]int{1: 1}},
		{Op: "_WAIT", Args: []string{"120"}, Nums: map[int]int{0: 120}},
		{Op: "_CLEAR4"},
		{Op: "_PRINT13", Args: []string{"ALLTIME_TEXT", "338"}, Nums: map[int]int{1: 338}},
		{Op: "_WAIT", Args: []string{"40"}, Nums: map[int]int{0: 40}},
		{Op: "_CLEAR4"},
		{Op: "_FLASHOFF", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_MATRIXLGT", Args: []string{"0"}, Nums: map[int]int{0: 0}},
		{Op: "_PRINT13", Args: []string{"HI_1", "336"}, Nums: map[int]int{1: 336}},
		{Op: "_PRINT13_NUMBER", Args: []string{"DEMO_HI_SCORE_0", "368"}, Nums: map[int]int{1: 368}},
		{Op: "_PRINT13", Args: []string{"DEMO_HI_NAME_0", "344"}, Nums: map[int]int{1: 344}},
		{Op: "_FLASHON", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_WAIT", Args: []string{"140"}, Nums: map[int]int{0: 140}},
		{Op: "_CLEAR4"},
		{Op: "_FLASHOFF", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_MATRIXLGT", Args: []string{"0"}, Nums: map[int]int{0: 0}},
		{Op: "_PRINT13", Args: []string{"HI_2", "336"}, Nums: map[int]int{1: 336}},
		{Op: "_PRINT13_NUMBER", Args: []string{"DEMO_HI_SCORE_1", "368"}, Nums: map[int]int{1: 368}},
		{Op: "_PRINT13", Args: []string{"DEMO_HI_NAME_1", "344"}, Nums: map[int]int{1: 344}},
		{Op: "_FLASHON", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_WAIT", Args: []string{"140"}, Nums: map[int]int{0: 140}},
		{Op: "_CLEAR4"},
		{Op: "_FLASHOFF", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_MATRIXLGT", Args: []string{"0"}, Nums: map[int]int{0: 0}},
		{Op: "_PRINT13", Args: []string{"HI_3", "336"}, Nums: map[int]int{1: 336}},
		{Op: "_PRINT13_NUMBER", Args: []string{"DEMO_HI_SCORE_2", "368"}, Nums: map[int]int{1: 368}},
		{Op: "_PRINT13", Args: []string{"DEMO_HI_NAME_2", "344"}, Nums: map[int]int{1: 344}},
		{Op: "_FLASHON", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_WAIT", Args: []string{"140"}, Nums: map[int]int{0: 140}},
		{Op: "_CLEAR4"},
		{Op: "_FLASHOFF", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_MATRIXLGT", Args: []string{"0"}, Nums: map[int]int{0: 0}},
		{Op: "_PRINT13", Args: []string{"HI_4", "336"}, Nums: map[int]int{1: 336}},
		{Op: "_PRINT13_NUMBER", Args: []string{"DEMO_HI_SCORE_3", "368"}, Nums: map[int]int{1: 368}},
		{Op: "_PRINT13", Args: []string{"DEMO_HI_NAME_3", "344"}, Nums: map[int]int{1: 344}},
		{Op: "_FLASHON", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_WAIT", Args: []string{"140"}, Nums: map[int]int{0: 140}},
		{Op: "_CLEAR4"},
		{Op: "_FLASHOFF", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_CLEAR4"},
		{Op: "_PRINT5", Args: []string{"PLAYERSTEXT", "336"}, Nums: map[int]int{1: 336}},
		{Op: "_PRINT5", Args: []string{"DEMO_BALLSTEXT", "1684"}, Nums: map[int]int{1: 1684}},
		{Op: "0"},
	}
}

func (d *demoCore) infoCommand(c presentation.Command) bool {
	if !d.infoPresentation {
		return false
	}
	for _, w := range demoInfoProgram() {
		if c.Op != w.Op || !reflect.DeepEqual(c.Args, w.Args) || !sameDemoNums(c.Nums, w.Nums) {
			continue
		}
		if c.Op == "_PRINT5" {
			return d.game.Display.Content.Fonts["5"] == d.gameplayFont && d.infoTextValid(c.Args[0])
		}
		if len(c.Args) > 0 && len(c.Args[0]) >= 8 && c.Args[0][:8] == "DEMO_HI_" && !d.infoFactoryValid() {
			return false
		}
		if c.Op == "_PRINT13" || c.Op == "_RULLGARDIN_UPP" || c.Op == "_RULLGARDIN_NED" {
			return d.game.Display.Content.Fonts["13"] == d.drainFont && d.infoTextValid(c.Args[0])
		}
		if c.Op == "_PRINT13_NUMBER" {
			if d.game.Display.Content.Fonts["13"] != d.drainFont {
				return false
			}
			if c.Args[0] == "JACKVALUE" {
				_, ok := demoScoreAbove(d.game.Jackpot, Number(0))
				return ok
			}
			return d.infoFactoryValid() && d.infoTextValid(c.Args[0])
		}
		return true
	}
	return false
}
func (d *demoCore) infoTextValid(label string) bool {
	want, ok := d.infoTexts[label]
	if !ok || len(want) == 0 {
		return false
	}
	actual := d.game.Display.Content.Texts[label]
	if label == "PLAY_TEXT" {
		if len(actual) != len(want) {
			return false
		}
		for i, v := range want {
			if i != 8 && i != 18 && actual[i] != v {
				return false
			}
		}
		return actual[8] == '8' && len(d.drainBallText) > 5 && actual[18] == d.drainBallText[5]
	}
	if label == "BONUS_TEXT" {

		if len(actual) != len(want) {
			return false
		}
		for i, v := range want {
			if i != 11 && actual[i] != v {
				return false
			}
		}
		return actual[11] == '8'
	}
	return bytes.Equal(actual, want)
}
func (d *demoCore) infoFactoryValid() bool {
	if sha256.Sum256(d.infoFactory[:]) != d.infoFactorySeal {
		return false
	}
	if d.factoryTopSource != "verified-native-factory-volatile" || d.factoryTop == nil || *d.factoryTop != Number(50000000) {
		return false
	}
	for rank, value := range []uint64{50000000, 25000000, 10000000, 5000000} {
		digits := Number(value)
		if !bytes.Equal(d.infoFactory[rank*16:rank*16+12], digits[:]) {
			return false
		}
		label := "DEMO_HI_SCORE_" + strconv.Itoa(rank)
		name := "DEMO_HI_NAME_" + strconv.Itoa(rank)
		if !bytes.Equal(d.game.Display.Content.Texts[label], d.infoFactory[rank*16:rank*16+12]) || !bytes.Equal(d.game.Display.Content.Texts[name], d.infoFactory[rank*16+12:rank*16+16]) {
			return false
		}
	}
	return true
}

// Preflight every mandatory operand before FJANTTEXT, mutable text, flash or
// matrix writes. SHOW_HI_ETC's spring guard precedes every presentation read.
func (d *demoCore) showInfoPreflight() error {
	if d.terminal != nil {
		return nil
	}
	if d.failure != nil {
		return d.failure
	}
	if !d.infoPresentation || !d.highScoreOperands || d.infoCount != 720 || d.game.matrix.active {
		return d.reject("SHOW_HI_ETC", "eligibility", "requires admitted idle threshold720")
	}
	if err := d.checkHighScorePreflight(); err != nil {
		return err
	}
	if d.game.Physics.SpringValid {
		return nil
	}
	if !d.infoFactoryValid() {
		return d.reject("SHOW_HI_ETC", "four factory records", "missing, changed or unverified volatile records")
	}
	for _, c := range demoInfoProgram() {

		if c.Op == "_RULLGARDIN_NED" && c.Args[0] == "PLAY_TEXT" {
			continue
		}
		if !d.command(c) {
			return d.reject("SHOW_HI_ETC", fmt.Sprintf("mandatory %s", c.Op), "presentation input unavailable before side effects")
		}
	}
	want := d.infoTexts["PLAY_TEXT"]
	actual := d.game.Display.Content.Texts["PLAY_TEXT"]
	if len(want) != 21 || len(actual) != len(want) || len(d.drainBallText) != 7 || !bytes.Equal(d.game.Display.Content.Texts["DEMO_BALLSTEXT"], d.drainBallText) {
		return d.reject("SHOW_HI_ETC", "PLAY_TEXT", "missing mutable source data")
	}
	for i, v := range want {
		if i != 8 && i != 18 && actual[i] != v {
			return d.reject("SHOW_HI_ETC", "PLAY_TEXT", "changed immutable text")
		}
	}
	return nil
}
func (d *demoCore) showInfo() error {
	if d.terminal != nil {
		return nil
	}
	if err := d.showInfoPreflight(); err != nil {
		return err
	}
	g := d.game
	if g.Physics.SpringValid {
		d.infoCount++
		g.Display.Score(g.Score.String())
		g.Display.FlushPrint(g.matrixNumber)
		return nil
	}
	d.fjantText = true
	g.Display.Content.Texts["PLAY_TEXT"][8] = '8'
	g.Display.Content.Texts["PLAY_TEXT"][18] = d.drainBallText[5]
	d.dotReady = false
	d.infoCount = 0
	if err := d.install(demoInfoProgram()); err != nil {
		return err
	}
	d.infoStarts++
	d.infoCount++
	g.Display.Score(g.Score.String())

	g.matrixDispatch()
	g.Display.FlushPrint(g.matrixNumber)
	return nil
}
