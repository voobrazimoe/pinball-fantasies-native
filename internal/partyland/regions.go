package partyland

// Original AREALISTA order is significant: first match wins. LASTCHECK is
// cleared outside every region; LASTAREA retains the preceding callback.
type region struct {
	x1, y1, x2, y2 int16
	label          string
}

var lower = []region{
	{90, 15, 130, 40, "BYGEL11"}, {200, 15, 240, 40, "BYGEL9"}, {160, 35, 182, 55, "GROPC"},
	{47, 122, 67, 146, "GROPD"}, {120, 165, 150, 185, "GROPB"}, {220, 280, 260, 300, "BYGEL14"},
	{280, 300, 320, 345, "CLOSE1"}, {260, 312, 268, 321, "GROPE"}, {25, 435, 35, 445, "BYGEL3"},
	{263, 435, 273, 445, "BYGEL4"}, {5, 455, 15, 465, "BYGEL1"}, {284, 455, 294, 465, "BYGEL2"},
	{305, 455, 320, 540, "BYGEL12"}, {308, 540, 320, 576, "BYGEL28"},
}
var upper = []region{
	{132, 40, 238, 51, "CLOSE2"}, {167, 6, 201, 40, "CLOSE2"}, {132, 5, 160, 50, "CLOSE2"},
	{40, 40, 75, 100, "BYGEL10"}, {132, 58, 148, 63, "BYGEL5"}, {222, 58, 238, 63, "BYGEL8"},
	{162, 64, 178, 69, "BYGEL6"}, {192, 64, 208, 69, "BYGEL7"}, {40, 80, 75, 120, "OPEN2"},
	{175, 100, 200, 130, "BYGEL4B"}, {260, 130, 280, 150, "BYGEL13"}, {3, 245, 22, 270, "GROPA"},
}

func (g *Game) checkAreas() {
	b := g.Physics.Ball
	x, y := uint16(b.PixelX+8), uint16(b.PixelY+8+g.Physics.ScreenOffset)
	list := lower
	if b.High {
		list = upper
	}
	if g.Physics.Tilted {
		list = []region{{120, 165, 150, 185, "GROPB"}, {160, 35, 182, 55, "GROPC_T"}, {47, 122, 67, 146, "GROPD_T"}}
		if b.High {
			list = nil
		}
	}
	for _, r := range list {
		if x >= uint16(r.x1) && x <= uint16(r.x2) && y >= uint16(r.y1) && y <= uint16(r.y2) {
			if g.lastCheck != r.label {
				g.lastCheck = r.label
				g.trigger(r.label)
				g.lastArea = g.lastCheck
			}
			return
		}
	}
	g.lastCheck = ""
}
func (g *Game) trigger(label string) {
	g.emit("Switch", label, 0)
	switch label {
	case "GROPC_T", "GROPD_T":
		g.Physics.SetBall(15, 47, 0, 0, false)
		g.startDrop()
	case "BYGEL1", "BYGEL2":
		// Original right lane also checks bit 39; palette entry 40 aliases 39.
		if g.Lights[39] {
			g.light(39, false)
			g.light(40, false)
			g.effect("EXTRABALL2", 10000, 5000)
			g.ExtraBalls++
			g.light(51, true)
		} else {
			g.sound("SBYGEL1")
			g.score("BCD50030", 50030)
		}
	case "BYGEL3", "BYGEL4":
		g.effect("BYGELSETB", 10040, 1000)
		g.sound("SBYGEL2")
	case "BYGEL12":
		g.SkillTime = 300
		g.Physics.SpringValid = false
		g.InhibitReverseTime = 120
	case "BYGEL28":
		g.Physics.SpringValid = true
	case "BYGEL14":
		g.InhibitReverseTime = 0
	case "CLOSE1":
		if g.lastArea == "BYGEL12" {
			g.inChute = false
			g.Session.SelectionOpen = false // CLOSE1 writes ADDPLAYERS=FALSE.
			g.partyFlash = false
			g.music("S_MAIN")
			g.Audio.ReturnPosition = 1
			g.beginMatrix("PARTY_OFFTS")
		}
	case "BYGEL10":
		g.gate(true)
	case "CLOSE2":
		if g.lastArea == "BYGEL10" {
			g.skyride()
		}
	case "BYGEL5":
		g.pukeLetter(1)
	case "BYGEL8":
		g.pukeLetter(2)
	case "BYGEL6":
		g.pukeLetter(4)
	case "BYGEL7":
		g.pukeLetter(5)
	case "BYGEL9":
		if g.lastArea == "BYGEL11" {
			g.loop()
		}
	case "BYGEL11":
		if g.lastArea == "BYGEL9" {
			g.reverse()
		}
	case "BYGEL13":
		g.cyclone()
	case "GROPA":
		g.snack()
	case "GROPB":
		g.arcade()
	case "GROPC":
		g.hidden()
	case "GROPD":
		g.tunnel()
	case "GROPE":
		g.dragon()
	}
}
