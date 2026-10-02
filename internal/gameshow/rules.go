package gameshow

import "strconv"

type rectangle struct {
	x1, y1, x2, y2 int16
	label          string
}

func (r rectangle) contains(x, y int16) bool {
	return uint16(x) >= uint16(r.x1) && uint16(x) <= uint16(r.x2) && uint16(y) >= uint16(r.y1) && uint16(y) <= uint16(r.y2)
}

var lowerAreas = []rectangle{{230, 280, 260, 330, "CLOSE1"}, {305, 512, 320, 576, "BYGEL28"}, {300, 400, 320, 450, "OPEN1"}, {102, 23, 122, 45, "BYGEL11"}, {190, 25, 210, 45, "BYGEL12"}, {103, 233, 122, 254, "GROPA"}, {1, 520, 20, 555, "GROPB"}, {1, 460, 25, 500, "CLOSE4"}, {25, 435, 35, 445, "BYGEL3"}, {263, 435, 273, 445, "BYGEL4"}, {5, 455, 15, 465, "BYGEL1"}, {284, 455, 294, 465, "BYGEL2"}, {120, 130, 130, 140, "BYGEL9"}, {222, 104, 232, 114, "BYGEL8"}, {90, 152, 100, 162, "BYGEL13"}}
var upperAreas = []rectangle{{1, 30, 20, 50, "BYGEL6"}, {45, 70, 70, 95, "CLOSE3"}, {31, 78, 51, 98, "BYGEL10"}, {290, 90, 320, 125, "BYGEL5"}, {155, 75, 185, 105, "BYGEL7"}, {222, 104, 232, 114, "BYGEL8"}, {120, 130, 130, 140, "BYGEL9"}, {1, 50, 20, 70, "BYGEL6B"}}

func (g *Game) checkAreas() {
	b := &g.Physics.Ball
	areas := lowerAreas
	if b.High {
		areas = upperAreas
	}
	if g.Physics.Tilted {
		areas = []rectangle{{1, 520, 20, 555, "GROPB"}}
		if b.High {
			areas = nil
		}
	}
	for _, r := range areas {
		if r.contains(b.PixelX+8, b.PixelY+8+g.Physics.ScreenOffset) {
			if r.label != g.lastCheck {
				g.lastCheck = r.label
				g.area(r.label)
				g.lastArea = r.label
			}
			return
		}
	}
	g.lastCheck = ""
}
func (g *Game) addMoney(loops bool) {
	if g.Special && g.MoneyMania && g.LoopsAndTraps == loops {
		n := uint64(500000)
		if loops {
			n = 1000000
		}
		g.MoneyTotal.Add(number(n))
	}
}
func (g *Game) cashAdd() { g.CashPot.Add(number(7130)) }
func (g *Game) jackAdd() { g.Jackpot.Add(number(100000)) }
func (g *Game) touch(i int) {
	g.addMoney(false)
	g.sound("S_TOUCH2")
	if i < 2 {
		mine, other := 2+i, 3-i
		if (i == 0 && g.Lights[other]) || (i == 1 && g.Lights[other]) {
			g.flash(2, 2)
			g.flash(3, 2)
			g.wait("FLASHA23", 60, func() { g.off(2); g.off(3) })
			g.effect("DOLLARTOUCH2")
			g.synced(18, true)
			g.Lights[18] = true
			g.gate(1, true)
		} else {
			g.flash(mine, 6)
			g.Lights[mine] = true
			site := "FLASHA2"
			if i == 1 {
				site = "FLASHA3"
			}
			g.wait(site, 25, func() { g.endFlash(mine); g.light(mine, true) })
			g.effect("DOLLARTOUCH")
		}
		return
	}
	n := i + 5
	if !g.Lights[n] {
		return
	}
	label := "TOUCHB"
	base := 7
	site := "UP_B"
	if i >= 4 {
		label = "TOUCHC"
		base = 9
		site = "UP_C"
	}
	g.effect(label)
	g.gate(i+1, true)
	g.light(n, false)
	if !g.Lights[base] && !g.Lights[base+1] {
		g.wait(site, 60, func() {
			g.sound("SBRICKUPP")
			g.light(base, true)
			g.light(base+1, true)
			g.gate(base-4, false)
			g.gate(base-3, false)
		})
	}
}
func (g *Game) area(label string) {
	g.emit("Area", label, 0)
	switch label {
	case "CLOSE1":
		if g.lastArea == "OPEN1" {
			g.gate(2, false)
			g.inChute = false
			g.Cue("S_MAIN")
			g.music.ReturnPosition = 1
			g.beginMatrix("PARTY_OFFTS")
		}
	case "BYGEL28":
		g.Physics.SpringValid = true
	case "OPEN1":
		g.Physics.SpringValid = false
		g.gate(2, true)
	case "CLOSE4":
		g.gate(8, false)
	case "CLOSE3":
	case "GROPA":
		g.cashCapture()
	case "GROPB":
		g.wheelCapture()
	case "BYGEL1", "BYGEL2":
		g.effect("BYGELSETA")
		g.soundVolume("SBYGEL1", 32)
	case "BYGEL3", "BYGEL4":
		g.effect("BYGELSETB")
		g.sound("SBYGEL2")
	case "BYGEL8":
		g.effect("BYGELSETF")
	case "BYGEL9":
		g.effect("BYGELSETG")
	case "BYGEL13":
		g.effect("BYGELSETK")
	case "BYGEL6B":
		g.jackAdd()
	case "BYGEL6":
		g.cashAdd()
		g.effect("BYGELSETD")
		old := g.timers[8]
		g.timers[8] = 600
		if old != 0 {
			g.effect("LOOPMILLION")
		}
		if g.Lights[17] {
			g.playJingle("SJINGLE21")
		}
		g.synced(4, false)
	case "BYGEL5":
		g.cashAdd()
		g.addMoney(true)
		g.effect("BYGELSETC")
		g.timers[6] = 240
		g.gate(7, true)
		if g.timers[1] != 0 {
			g.timers[1] = 1
			g.effect("JACKPOT")
			g.Jackpot = number(10000000)
			g.timers[0] = 300
			g.flash(5, 10)
			g.playJingle("SJINGLE3")
		}
		g.timers[3] = 240
	case "BYGEL7":
		g.clockwise()
	case "BYGEL11":
		g.effect("BYGELSETI")
		if g.lastArea == "BYGEL12" {
			if g.timers[4] != 0 && g.Prizes[3] == 0 {
				g.timers[4] = 0
				g.litPrize(3, false)
			} else if g.timers[9] != 0 && g.Prizes[4] == 0 {
				g.litPrize(4, true)
			}
		}
	case "BYGEL12":
		g.addMoney(true)
		g.effect("BYGELSETJ")
		if g.lastArea != "BYGEL11" {
			return
		}
		if g.Lights[11] {
			g.effect("EXTRA_BALL")
			g.off(11)
			g.light(31, true)
		}
		if g.timers[10] != 0 && g.Prizes[5] == 0 {
			g.timers[10] = 0
			g.litPrize(5, false)
		} else {
			g.gate(7, false)
			g.timers[2] = 240
			g.timers[5] = 240
		}
	case "BYGEL10":
		g.effect("BYGELSETH")
		if g.lastArea != "CLOSE3" {
			return
		}
		g.jackAdd()
		g.cashAdd()
		g.addMoney(true)
		g.anotherSkill()
		if g.timers[3] != 0 {
			if g.AllSix {
				return
			}
			if g.TopThree {
				g.wherePrize(3)
				return
			}
			if g.Prizes[0] == 0 {
				g.timers[3] = 0
				g.litPrize(0, false)
				return
			}
		}
		if g.timers[5] != 0 && !g.AllSix {
			if g.TopThree {
				g.wherePrize(4)
				return
			}
			g.timers[5] = 0
			if g.Prizes[1] == 0 {
				g.litPrize(1, true)
			}
		}
	}
}

var prizeLamps = [6]int{13, 14, 15, 28, 29, 30}
var prizeNames = [6]string{"TV", "TRIP", "CAR", "BOAT", "HOUSE", "PLANE"}

func (g *Game) litPrize(i int, invert bool) {
	g.Prizes[i] = 1
	g.effect(prizeNames[i] + "LIT")
	g.synced(prizeLamps[i], invert)
	base := 0
	if i >= 3 {
		base = 3
	}
	if g.Prizes[base] != 0 && g.Prizes[base+1] != 0 && g.Prizes[base+2] != 0 {
		g.synced(17, true)
		g.Lights[17] = true
		g.gate(1, true)
	}
}
func (g *Game) wherePrize(i int) {
	if g.Prizes[i] != 0 {
		return
	}
	g.beginMatrix("LITEPRIZE_RIGHTTS")
	if i == 5 {
		g.beginMatrix("LITEPRIZE_LEFTTS")
	}
	g.timers[[6]int{3, 5, 6, 4, 9, 10}[i]] = 600
}
func (g *Game) clockwise() {
	g.cashAdd()
	g.jackAdd()
	g.addMoney(true)
	g.effect("BYGELSETE")
	if g.timers[0] != 0 {
		g.effect("S_JACKPOT")
		g.timers[0] = 1
	}
	g.timers[7] = 660
	g.synced(12, true)
	if g.timers[6] != 0 && g.Prizes[2] < 1 {
		g.timers[6] = 0
		g.litPrize(2, false)
		return
	}
	if g.Prizes[5] == 0 && g.timers[6] != 0 && g.TopThree && !g.AllSix {
		g.timers[6] = 0
		g.wherePrize(5)
		return
	}
	if g.timers[2] != 0 {
		if g.Multiplier == 10 {
			return
		}
		values := []uint8{1, 2, 3, 4, 6, 8, 10}
		index := 0
		for values[index] != g.Multiplier {
			index++
		}
		g.Multiplier = values[index+1]
		if g.music.Priority <= g.Display.Jingle("SJINGLE2").Priority {
			g.beginMatrix("_BONUSX" + strconv.Itoa(int(g.Multiplier)) + "TS")
			g.playJingle("SJINGLE2")
		}
		g.light(33+index, true)
		return
	}
	if g.timers[6] != 0 {
		g.RaisingMillions[5]++
		g.effect("RAISING_M")
	}
}
func (g *Game) anotherSkill() {
	g.Skills++
	if g.Skills == 1 {
		g.Skills++
	}
	if g.Skills%6 == 0 {
		n := g.Skills / 6
		if n == 2 {
			g.playJingle("SJINGLE19")
			g.synced(11, true)
			g.Lights[11] = true
			return
		}
		g.LoopsAndTraps = n > 2 && n%2 == 1
		label := "MONEYMANIA"
		if g.LoopsAndTraps {
			label = "MONEYMANIA2"
		}
		g.effect(label)
		g.Special = true
		g.MoneyMania = true
		g.light(32, true)
		return
	}
	if g.Skills/6 == 1 {
		g.effect("SKILLSHOT_XB")
	} else {
		if g.Skills/6 > 1 {
			g.writeSkillText((g.Skills/6 + 1) * 6)
		}
		g.effect("SKILLSHOT")
	}
}

// Source DOWNCOUNT jumps directly to its label on expiry. Several labels RETN,
// deliberately leaving later counters untouched for that interrupt.
func (g *Game) updateCounters() {
	dec := func(i int) bool {
		if g.timers[i] == 0 {
			return false
		}
		g.timers[i]--
		return g.timers[i] == 0
	}
	if dec(0) {
		g.off(5)
		return
	}
	if g.timers[0] == 120 {
		g.flash(5, 2)
	}
	if dec(1) {
		g.off(16)
		return
	}
	if g.timers[1] == 120 {
		g.flash(16, 2)
	}
	g.syncFlasher = (g.syncFlasher + 1) % 20
	if g.spinCounter != 0 {
		g.spinCounter--
		if g.spinCounter == 0 {
			g.nextSpin()
			return
		}
	}
	for _, i := range []int{2, 3, 4, 5, 6} {
		if dec(i) {
			return
		}
	}
	if dec(7) {
		g.off(12)
		return
	}
	if g.timers[7] == 120 {
		g.flash(12, 2)
	}
	if dec(8) {
		g.off(4)
		return
	}
	if g.timers[8] == 120 {
		g.flash(4, 3)
	}
	// SHOW does not DOWNCOUNT HOUSEcounter or PLANEcounter.
}
func (g *Game) cashCapture() {
	g.Physics.Ball.Hold = true
	if g.Special {
		g.light(6, true)
		g.Physics.SetBall(103, 233, 0, 0, false)
		g.wait("SLAPP2", 30, g.ejectCash)
		return
	}
	if g.AllSix {
		g.light(6, true)
		g.synced(27, false)
		g.effect("SHOOTTHEBALL")
		g.music.ReturnPosition = 0
		g.Physics.SetBall(304, 535, 10, 0, false)
		g.Physics.Ball.Hold = false
		g.music.Priority = 0
		g.gate(1, true)
		g.BillionEnabled = true
		return
	}
	g.jackAdd()
	if g.timers[7] == 0 {
		g.effect("CASHPOT")
	} else {
		g.timers[7] = 10
		for i := 0; i < 5; i++ {
			g.CashPot5.Add(g.CashPot)
		}
		g.effect("CASHPOT5")
	}
	g.Physics.SetBall(103, 233, 0, 0, false)
	g.wait("RELEASE_BALL", 160, func() { g.light(6, true); g.wait("BASIL_FAWLTY", 40, g.ejectCash); g.CashPot5 = Decimal{} })
}
func (g *Game) ejectCash() {
	g.light(6, false)
	g.sound("SGROP")
	g.Physics.SetBall(103, 233, 100, 1700, false)
	g.Physics.Ball.Hold = false
}
func (g *Game) ejectWheel() {
	g.ScreenForce = -1
	g.Physics.TargetRaster = -1
	g.Physics.Ball.Hold = false
	g.Physics.Ball.VY = -3500
	g.off(18)
	g.beginMatrix("RENSATS")
}

var spinTimes = []uint16{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 6, 7, 7, 7, 7, 8, 8, 8, 9, 10, 10, 10, 10, 11, 11, 11, 11, 12, 12, 14, 16, 19, 22, 25, 50}
var spinScores = []uint64{25000, 50000, 100000, 250000, 500000, 1000000, 2500000, 5000000}

func (g *Game) wheelCapture() {
	g.gate(8, true)
	g.Physics.Ball.Hold = true
	g.Physics.SetBall(4, 529, 0, 0, false)
	if g.Special || g.Physics.Tilted {
		g.wait("SLAPP_HONOM", 30, func() { g.Physics.Ball.VY = -3500; g.sound("SNEWBALL"); g.Physics.Ball.Hold = false })
		return
	}
	if g.BillionEnabled {
		g.BillionEnabled = false
		g.effect("BILLION")
		g.flash(27, 4)
		g.wait("TURNOFF_BILLION", 250, func() {
			for _, n := range []int{27, 13, 14, 15, 28, 29, 30} {
				g.off(n)
			}
			g.Prizes = [6]uint8{}
			g.TopThree = false
			g.AllSix = false
			g.Physics.Ball.VY = -3500
			g.sound("SNEWBALL")
			g.Physics.Ball.Hold = false
		})
		return
	}
	if !g.Lights[17] {
		g.gate(1, false)
	}
	g.playJingle("S_MYSTERY")
	g.spinIndex = 0
	g.spinCounter = spinTimes[0]
	g.beginMatrix("RENSA2TS")
	g.light(19+g.spinLight, false)
	v := uint8(g.clock)
	if g.Lights[17] {
		for _, pair := range [][2]int{{0, 0}, {1, 1}, {2, 2}, {3, 6}, {4, 5}, {5, 4}} {
			if g.Prizes[pair[0]] == 1 {
				v = uint8(pair[1])
				break
			}
		}
	}
	g.spinLight = int((v - 2) & 7)
	g.spinning = true
	if g.Physics.Configured {
		g.Physics.TargetRaster = 220
		if g.Physics.Settings.Resolution == 0 {
			g.Physics.TargetRaster = 270
		}
	} else {
		g.ScreenForce = 187
	}
}
func (g *Game) nextSpin() {
	g.spinIndex++
	if g.spinIndex == len(spinTimes) {
		g.spinning = false
		g.beginMatrix("FLASHMATRIXTS")
		if g.Lights[17] {
			g.task(func() bool { g.winPrize(); return true })
		} else {
			g.wait("END_OF_SPIN", 100, func() { g.score("END_OF_SPIN", g.spinScore); g.ejectWheel() })
		}
		return
	}
	g.spinCounter = spinTimes[g.spinIndex]
	g.light(19+g.spinLight, false)
	g.spinLight = (g.spinLight + 1) & 7
	g.light(19+g.spinLight, true)
	g.spinScore = spinScores[g.spinLight]
	g.beginMatrix("SPINTS")
}
func (g *Game) winPrize() {
	base := 0
	if g.TopThree {
		base = 3
	}
	i := base
	for i < base+2 && g.Prizes[i] == 2 {
		i++
	}
	g.endFlash(prizeLamps[i])
	g.light(prizeLamps[i], true)
	g.Prizes[i] = 2
	g.music.Priority = 0
	g.effect("YOUWIN" + prizeNames[i])
	if i == 2 {
		g.off(17)
		g.synced(16, false)
		g.Lights[16] = true
		g.timers[1] = 1500
		g.TopThree = true
		g.gate(1, false)
	}
	if i == 5 {
		g.off(17)
		g.AllSix = true
		g.gate(1, false)
	}
}

// SHOW another_skill -> Put_In_Text. Only a zero hundreds digit is blanked;
// AX>=100 advances BX once, retaining the previous byte at SKILLTEXT[16].
func (g *Game) writeSkillText(value uint16) {
	at := 16
	if value >= 100 {
		at++
	}
	buf := g.Display.MutableText("SKILLTEXT", 20)
	for i := 2; i >= 0; i-- {
		digit := byte(value % 10)
		value /= 10
		buf[at+i] = digit + '7'
		if i == 0 && digit == 0 {
			buf[at+i] = '*'
		}
	}
	// At >=100 the final digit replaces the DB terminator, exactly as source.

	g.Display.Content.Texts["SKILLTEXT"] = buf
}
