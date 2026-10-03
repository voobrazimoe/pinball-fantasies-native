package speeddevils

var lowerAreas = []rectangle{
	{260, 35, 275, 60, "GROPA"}, {295, 50, 320, 70, "CLOSE1"}, {140, 50, 170, 70, "BYGEL16"}, {10, 65, 40, 95, "BYGEL9"}, {140, 90, 160, 110, "BYGEL17"},
	{253, 124, 263, 130, "BYGEL7"}, {228, 132, 238, 138, "BYGEL6"}, {205, 140, 215, 146, "BYGEL5"}, {25, 310, 35, 320, "BYGEL8"},
	{25, 435, 35, 445, "BYGEL3"}, {263, 435, 273, 445, "BYGEL4"}, {5, 455, 15, 465, "BYGEL1"}, {284, 455, 294, 465, "BYGEL2"},
	{300, 480, 320, 500, "OPEN2"}, {305, 512, 320, 576, "BYGEL28"},
}
var upperAreas = []rectangle{{145, 13, 180, 35, "BYGEL10"}, {295, 50, 320, 60, "NEDSLAPP"}, {120, 64, 150, 78, "BYGEL12"}, {1, 250, 20, 290, "BYGEL11"}}

func (g *Game) checkAreas() {
	b := &g.Physics.Ball
	if g.Physics.Tilted {
		return
	}
	areas := lowerAreas
	if b.High {
		areas = upperAreas
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
	// CHECK_AREAS retains LASTAREA while outside a rectangle.
}
func (g *Game) area(label string) {
	g.emit("Area", label, 0)
	switch label {
	case "GROPA":
		g.pitstop()
	case "CLOSE1":
		if g.lastArea == "NEDSLAPP" {
			g.inChute = false
			g.Session.SelectionOpen = false // CLOSE1 writes ADDPLAYERS=FALSE.
			g.partyFlash = false
			g.Cue("S_MAIN")
			g.music.ReturnPosition = 2
			g.Physics.SpringValid = false
			g.beginMatrix("PARTY_OFFTS")
		}
	case "BYGEL16":
		g.jumpLane()
	case "BYGEL9":
		g.offroadLane()
	case "BYGEL7":
		g.pit(2)
	case "BYGEL6":
		g.pit(1)
	case "BYGEL5":
		g.pit(0)
	case "BYGEL8":
		g.effect("BYGELSETD")
	case "BYGEL3", "BYGEL4":
		g.sound("SBYGEL2")
		g.effect("BYGELSETB")
	case "BYGEL1", "BYGEL2":
		g.effect("BYGELSETA")
	case "OPEN2":
		g.Physics.SpringValid = false
	case "BYGEL28":
		g.Physics.SpringValid = true
	case "BYGEL10":
		g.loop(true)
	case "BYGEL11":
		g.loop(false)
	case "NEDSLAPP":
		g.Physics.Ball.VY = 0
		g.Physics.Ball.High = false
	}
}
