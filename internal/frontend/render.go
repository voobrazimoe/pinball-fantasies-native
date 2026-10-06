package frontend

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/settings"
	"strconv"
)

type View struct {
	SpeedMatrix    *presentation.Display
	GameshowMatrix *presentation.Display
	StonesMatrix   *presentation.Display
	Art            *assets.FrontendArt
	Font5, Font13  []byte
	Matrix         *presentation.Display
}

func NewView(a *assets.FrontendArt, table []byte) *View {
	return &View{Art: a, Font5: table[0x20450 : 0x20450+36*5], Font13: table[0x1ff40 : 0x1ff40+42*13], Matrix: presentation.New(1, table)}
}
func black(w, h int) *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(im, im.Rect, image.NewUniform(color.RGBA{0, 0, 0, 255}), image.Point{}, draw.Src)
	return im
}
func blit(dst *image.RGBA, p *assets.FrontendPicture, x, y, w, h, sx, sy int) {
	draw.Draw(dst, image.Rect(x, y, x+w, y+h), p.Frame(), image.Pt(sx, sy), draw.Src)
}
func (v *View) Frame(m *Model) *image.RGBA { return v.framePresentation(m, false) }
func (v *View) framePresentation(m *Model, full bool) *image.RGBA {
	c := m.Settings
	if full {
		c.ScrollMode = settings.ScrollOff
	}
	switch m.Mode {
	case Startup:
		return v.startup(m)
	case Options:
		return v.optionsFrame(m)
	case Selector, SelectorText:
		return v.selectorFrame(m)
	case Quit:
		return black(640, 240)
	}
	out := m.Session.Frame()
	draw.Draw(out, image.Rect(0, c.MatrixY(), 320, c.MatrixY()+33), image.NewUniform(color.RGBA{0, 0, 0, 255}), image.Point{}, draw.Src)
	switch m.Mode {
	case Playing:
		return m.Session.Frame()
	case GameEnd:
		if m.sourceScoreStage != 0 {
			return m.Session.Frame()
		}
		// end_gamen's six-sync handoff clears the display before score entry.
	case Paused:
		v.matrixText(out, "GAME PAUSED", 72, 2, m.Selected, c.MatrixY())
	case QuitQuestion:
		v.matrixText(out, "REALLY QUIT (Y OR N)", 0, 2, m.Selected, c.MatrixY())
	case Initials, EntryWait:
		if m.sourceScoreStage != 0 {
			return m.Session.Frame()
		}
		text := "HIGHSCORE PL " + strconv.Itoa(max(1, m.scorePlayer)) + " (" + string(m.Entry[:]) + ")"
		if m.Mode == EntryWait && m.Counter <= 30 {
			text = "********************"
		}
		v.matrixText(out, text, 0, 2, m.Selected, c.MatrixY())

	case TableAttract:
		if session, ok := m.Session.(interface{ AttractFrame(int) *image.RGBA }); ok {
			out = session.AttractFrame(m.Counter)
		}
		draw.Draw(out, image.Rect(0, c.MatrixY(), 320, c.MatrixY()+33), image.NewUniform(color.RGBA{0, 0, 0, 255}), image.Point{}, draw.Src)
		d := v.Matrix
		on, off := byte(242), byte(96)
		if m.Selected == 2 && v.SpeedMatrix != nil {
			d = v.SpeedMatrix
			on, off = 128, 98
		}
		if m.Selected == 3 && v.GameshowMatrix != nil {
			d = v.GameshowMatrix
			on, off = 153, 114
		}
		if m.Selected == 4 && v.StonesMatrix != nil {
			d = v.StonesMatrix
			on, off = 79, 231
		}
		var names, scores [4]string
		for i, e := range m.Scores[m.Selected-1] {
			names[i] = string(e.Name[:])
			scores[i] = e.Digits.String()
		}
		matrix := d.Attract(m.Counter, names, scores)
		if m.cheatTimeline != nil {
			matrix = m.cheatTimeline.Display()
		}
		if m.End == Completed && m.cheatTimeline == nil && m.gameOverTimeline == nil {
			players := []string{m.Final.String()}
			if len(m.scoreQueue) > 0 {
				players = make([]string, len(m.scoreQueue))
				for i, score := range m.scoreQueue {
					players[i] = score.String()
				}
			}
			matrix = d.GameOverPlayers(m.Counter, players, names, scores)
		}
		if m.gameOverTimeline != nil && m.cheatTimeline == nil {
			matrix = m.gameOverTimeline.Display()
		}
		var p [768]byte
		// Both table palettes use 20-DAC gray for their unlit lattice.
		p[int(off)*3], p[int(off)*3+1], p[int(off)*3+2] = assets.DACRGB(20), assets.DACRGB(20), assets.DACRGB(20)
		p = presentation.MatrixPaletteMode(p, 0, on)
		if m.Selected == 4 {
			p = d.SourceMatrixPalette(p)
		}
		matrix.DrawAt(out, p, off, on, c.MatrixY())

	}
	return out
}
func (v *View) matrixText(dst *image.RGBA, s string, x, y, table int, matrixY int) {
	base := v.Matrix
	on, off := byte(242), byte(96)
	if table == 2 && v.SpeedMatrix != nil {
		base, on, off = v.SpeedMatrix, 128, 98
	}
	if table == 3 && v.GameshowMatrix != nil {
		base, on, off = v.GameshowMatrix, 153, 114
	}
	if table == 4 && v.StonesMatrix != nil {
		base, on, off = v.StonesMatrix, 79, 231
	}
	d := *base
	d.Clear()
	d.On = true
	d.Text(s, x/2, y/2, 13)
	var p [768]byte
	p[int(off)*3], p[int(off)*3+1], p[int(off)*3+2] = assets.DACRGB(20), assets.DACRGB(20), assets.DACRGB(20)
	p = presentation.MatrixPaletteMode(p, 0, on)
	if table == 4 {
		p = d.SourceMatrixPalette(p)
	}
	d.DrawAt(dst, p, off, on, matrixY)
}
func (v *View) smallText(dst *image.RGBA, s string, x, y int) { v.dotText(dst, s, x, y, 5, 1, v.Font5) }
func (v *View) dotText(dst *image.RGBA, s string, x, y, h, scale int, font []byte) {
	for _, c := range s {
		idx := -1
		if c >= '0' && c <= '9' {
			idx = int(c - '0')
		} else if c >= 'A' && c <= 'Z' {
			idx = int(c-'A') + 10
		}
		if h == 13 {
			switch c {
			case '?':
				idx = 36
			case '(':
				idx = 37
			case ')':
				idx = 38
			case '-':
				idx = 39
			case '.':
				idx = 40
			case '*':
				idx = 41
			}
		}
		if idx >= 0 {
			for row := 0; row < h; row++ {
				for col := 0; col < 8; col++ {
					if font[idx*h+row]&(128>>uint(col)) != 0 {
						for yy := 0; yy < scale; yy++ {
							for xx := 0; xx < scale; xx++ {
								dst.SetRGBA(x+col*scale+xx, y+row*scale+yy, color.RGBA{242, 176, 64, 255})
							}
						}
					}
				}
			}

		}
		x += 8 * scale
	}
}

// PUTCHAR/fontlist: original 32-pixel-wide source cells, 18-pixel advance,
// second alphabet row at y=13. Unsupported spaces leave the backdrop intact.
func (v *View) selectorTextFont(dst *image.RGBA, f *assets.FrontendPicture, s string, x, y, level int) {
	v.selectorTextFontFade(dst, f, s, x, y, level, 20)
}
func (v *View) selectorTextFontFade(dst *image.RGBA, f *assets.FrontendPicture, s string, x, y, level, denominator int) {
	for _, c := range s {
		sx, sy := -1, 0
		switch {
		case c >= '0' && c <= '9':
			sx = int(c-'0') * 32
		case c >= 'A' && c <= 'J':
			sx = (10 + int(c-'A')) * 32
		case c >= 'K' && c <= 'Z':
			sx = int(c-'K') * 32
			sy = 14
		case c == '-':
			sx = 576
			sy = 14
		case c == '.':
			sx = 512
			sy = 14
		case c == '>':
			sx = 608
			sy = 14
		case c == ':':
			sx = 544
			sy = 14
		}
		if sx >= 0 {
			for yy := 0; yy < 14; yy++ {
				for xx := 0; xx < 24; xx++ {
					px := sx + xx
					py := sy + yy
					if px >= f.Width || py >= f.Height {
						continue
					}
					idx := f.Indices[py*f.Width+px]
					if idx != 0 {
						p := int(idx) * 3
						// SHOWTEXT/JULIUS installs the Party preview DAC bank;
						// the font's own CMAP is not the displayed palette.
						pal := v.Art.Tables[0].Palette
						c := color.RGBA{pal[p], pal[p+1], pal[p+2], 255}
						if f == v.Art.MonoFont {
							c = monoColorScale(idx, level, denominator)
						}
						dst.SetRGBA(x+xx, y+yy, c)
					}
				}
			}
		}
		x += 18
	}
}

var rasterGroups = func() [][]int {
	var groups [][]int
	for cnt := -49; cnt < 111; cnt += 8 {
		var g []int
		for k := 0; k < 8; k++ {
			line := cnt + k*7
			if line >= 0 && line < 95 {
				g = append(g, line)
			}
		}
		groups = append(groups, g)
	}
	return groups
}()

func (v *View) selectorFrame(m *Model) *image.RGBA {
	out := black(640, 240)
	// SHOWPICS/julius replaces both DAC banks. The sidebar shares the active
	// raster bank too; its own IFF palette is no longer the displayed palette.
	for y := 0; y < 240; y++ {
		bank := m.Page * 2
		if y >= 120 {
			bank++
		}
		if m.Mode == SelectorText {
			// RASTRACLEAR retains the preceding preview banks. SHOWTEXT then
			// calls JULIUS with table slot6 (Party) and disables the split.
			bank = m.PreviousPage * 2
			if y >= 120 {
				bank++
			}
			if m.TextTick >= 20 {
				bank = 0
			}
		}
		p := v.Art.Tables[bank].Palette
		for x := 0; x < 128; x++ {
			idx := int(v.Art.Logo.Indices[y*v.Art.Logo.Width+x]) * 3
			out.SetRGBA(x, y, color.RGBA{p[idx], p[idx+1], p[idx+2], 255})
		}
	}
	v.sidebar(out, m.SidebarTick)
	if m.Mode == Selector || m.TextTick < 20 {
		visible := [95]bool{}
		page := m.Page
		if m.Mode == Selector {
			groups := len(rasterGroups)
			if m.Reveal < 20 {
				groups = 0
			} else if m.Reveal < 20+groups {
				groups = m.Reveal - 20
			}
			for _, g := range rasterGroups[:groups] {
				for _, line := range g {
					visible[line] = true
				}
			}
		} else {
			page = m.PreviousPage
			for i := range visible {
				visible[i] = true
			}
			for _, g := range rasterGroups[:m.TextTick] {
				for _, line := range g {
					visible[94-line] = false
				}
			}
		}
		for i := 0; i < 2; i++ {
			pic := v.Art.Tables[page*2+i].Frame()
			for line := 0; line < 95; line++ {
				if !visible[line] {
					continue
				}
				y := line
				if i == 1 {
					y = 94 - line
				}
				draw.Draw(out, image.Rect(160, 10+i*125+y, 600, 11+i*125+y), pic, image.Pt(0, y), draw.Src)
			}
		}
	} else {
		lines := v.textPage(m)
		y := 14
		high := highTextPage(m.TextPage)
		font := v.Art.Font
		level := 63
		if m.TextTick < 45 {
			font = v.Art.MonoFont
			level = m.TextTick - 25
			if level < 0 {
				level = 0
			}
		} else if m.Counter == 0 {
			font = v.Art.MonoFont
			level = 19 - m.TextExitTick
			if level < 0 {
				level = 0
			}
		}
		if high {
			logo := v.Art.HighLogo
			if font == v.Art.MonoFont {
				logo = v.Art.HighMono
			} else {
				copy := *logo
				copy.Palette = v.Art.Tables[0].Palette
				logo = &copy
			}
			blit(out, logo, 184, 0, 400, 40, 0, 0)
			if font == v.Art.MonoFont {
				for yy := 0; yy < 40; yy++ {
					for xx := 0; xx < 400; xx++ {
						idx := logo.Indices[yy*logo.Width+xx]
						out.SetRGBA(184+xx, yy, monoColor(idx, level))
					}
				}
			}
			y += 10
		}
		for i, s := range lines {
			v.selectorTextFont(out, font, s, 164+(24-len(s))*18/2, y+i*18, level)
		}
	}
	// INTRO/creatretf writes CRTC start-address words 16,15,...,0.
	// Each planar address covers eight physical selector pixels. The 640px
	// scanline wraps into the next row, exactly as the original VRAM window.
	if m.SidebarTick <= 16 {
		shift := (16 - m.SidebarTick) * 8 * 4
		pan := black(640, 240)
		copy(pan.Pix, out.Pix[shift:])
		out = pan
	}
	return out
}

func highTextPage(page int) bool {
	return page == 1 || page == 2 || page == 5 || page == 6 || page == 9 || page == 10
}

func (v *View) textPage(m *Model) []string {
	if highTextPage(m.TextPage) {
		var lines []string
		names := []string{"PARTY LAND", "SPEED DEVILS", "BILLION DOLLAR", "STONES N BONES"}
		// WAITEND toggles BANPEK but selects the HITEXT of the old pair.
		for t := m.PreviousPage * 2; t < m.PreviousPage*2+2; t++ {
			lines = append(lines, "", fmt.Sprintf("     %-19s", names[t]))
			for rank, e := range m.Scores[t] {
				lines = append(lines, fmt.Sprintf("  %d. %s - %12d ", rank+1, e.Name, e.Digits.Uint64()))
			}
		}

		return lines
	}
	switch m.TextPage {
	case 3:
		return []string{"PINBALL FANTASIES PC", "BY FRONTLINE DESIGN", "", "PROGRAMMING BY:", "DANIEL FORSGREN", "GABRIEL BERGQVIST", "JOHAN LUNDMARK", "", "PRODUCED BY:", "BARRY SIMPSON", "AND", "STEWART GILRAY"}
	case 4:
		return []string{"", "", "ORIGINAL AMIGA GAME BY", "DIGITAL ILLUSIONS", "", "ANDREAS AXELSSON", "MARKUS NYSTROM", "OLOF GUSTAVSSON", "ULF MANDORFF", "FREDRIK LILIEGREN", "", ""}
	case 7:
		return []string{"SELECTOR INSTRUCTIONS", "", "F1 - PARTYLAND     ", "F2 - SPEED DEVILS  ", "F3 - BILLION DOLLAR", "F4 - STONES N BONES", "F5 - OPTIONS MENU  ", "", "SPACE TO SKIP", "ESC TO QUIT", "", ""}
	default:
		return []string{"", "", "", "", "-WINNERS-", "-DO NOT USE-", "-DRUGS-", "", "", "", "", ""}
	}
}
func (v *View) startup(m *Model) *image.RGBA {
	p := startupParts[m.Segment]
	w, h := 320, 240
	if p.picture == 4 {
		w, h = 640, 480
	}
	out := black(w, h)
	a := v.Art.Startup
	switch p.picture {
	case 0:
		blit(out, a[0], 0, 0, 320, 240, 0, 0)
		blit(out, a[1], 0, v.Art.StartupLowerY, 320, 240-v.Art.StartupLowerY, 0, 0)
	case 1:
		// INTRO loads Viking at row240, but sets CRTC start to row247.
		// The second IFF is 130 rows, starting at buffer row365.
		blit(out, a[3], 0, -7, 320, 126, 0, 0)
		blit(out, a[2], 0, 118, 320, 130, 0, 0)
	case 2:
		blit(out, a[4], 0, 0, 320, 110, 0, 0)
		blit(out, a[5], 0, 110, 320, 130, 0, 0)
	case 3:
		blit(out, a[6], 0, 0, 320, 200, 0, 0)
	case 4:
		// BIOS mode12h clears VRAM to index0, then INTRO fades the image's
		// entire DAC palette. Unwritten pixels therefore use its color0 too.
		pal := a[7].Palette
		draw.Draw(out, out.Rect, image.NewUniform(color.RGBA{pal[0], pal[1], pal[2], 255}), image.Point{}, draw.Src)
		blit(out, a[7], 0, 150, 640, 178, 0, 0)
	}
	fade := p
	phase := m.SegmentTick
	if fade.frames == 0 {
		for i := m.Segment - 1; i >= 0; i-- {
			q := startupParts[i]
			if q.picture == p.picture && q.frames > 0 {
				fade = q
				phase = q.frames - 1
				break
			}
		}
	}
	endpoint := func(level, value int) int {
		switch level {
		case 0:
			return 0
		case 127:
			return 63
		default:
			return value
		}
	}
	for i := 0; i < len(out.Pix); i += 4 {
		for c := 0; c < 3; c++ {
			value := int(out.Pix[i+c] >> 2)
			if fade.frames > 0 {
				if phase >= fade.frames {
					phase = fade.frames - 1
				}
				old, new := endpoint(fade.from, value), endpoint(fade.to, value)
				value = (old*(fade.frames-phase) + new*phase) / fade.frames
			} else {
				value = endpoint(p.to, value)
			}
			out.Pix[i+c] = assets.DACRGB(byte(value))
		}
	}
	return out
}

// ShowHighsTS/showithi dispatch durations. CLEAR4=5, CLEAR2=17,
// CLEAR3=81; scroll byte advances every four 71 Hz syncs.
func attractPanel(tick int, s Scores) (string, string, string) {
	return attractPanelForTable(tick, s, 1)
}

func attractPanelForTable(tick int, s Scores, table int) (string, string, string) {
	type panel struct {
		duration           int
		label, name, score string
	}
	parts := []panel{{5, "", "", ""}, {1, "", "", ""}, {121, "PINBALL FANTASIES", "", ""}, {18, "", "", ""}, {16, "THE REAL SIMULATOR", "", ""}, {90, "THE REAL SIMULATOR", "", ""}, {6, "", "", ""}, {91, "PARTY LAND", "", ""}, {82, "", "", ""}, {1, "", "", ""}}
	if table == 2 {
		parts[7].label = "SPEED DEVILS" // SDEV PL_TEXT, same ShowHighsTS timing.
	}
	scroll1 := "                     ADD PLAYERS WITH F1 TO F8 OR ENTER                      WINNERS DO NOT USE DRUGS                     "
	parts = append(parts, panel{(len(scroll1)-20)*4 + 1, scroll1, "", ""}, panel{40, "ALL TIME HIGHSCORES", "", ""}, panel{6, "", "", ""})
	for i, e := range s {
		parts = append(parts, panel{145, fmt.Sprintf("HIGHSCORE %d", i+1), string(e.Name[:]), e.Digits.String()}, panel{6, "", "", ""})
	}
	parts = append(parts, panel{29, "FRONTLINE DESIGN", "", ""}, panel{40, "FRONTLINE DESIGN", "", ""})
	scroll2 := "                     FLIP WITH LEFT AND RIGHT ALT SHIFT OR CTRL KEYS                     PUSH TABLE WITH SPACE                     CONTROL SPRING WITH DOWN ARROW KEY    ESC EXITS                     WINNERS DO NOT USE DRUGS                         "
	if table == 2 {
		// SDEV SCROLL_TEXT2 differs from PLAND's instructions.
		pad := "                     "
		scroll2 = pad + "FLIP WITH LEFT AND RIGHT ALT SHIFT OR CTRL KEYS" + pad + "PUSH TABLE WITH SPACE" + pad + "CONTROL SPRING WITH DOWN ARROW KEY OR MOUSE" + pad + "WINNERS DO NOT USE DRUGS" + pad + "P PAUSES GAME AND M TOGGLES INGAME MUSIC ON AND OFF" + pad + "ESC EXITS AND SAVES NEW HIGHSCORES" + pad
	}
	parts = append(parts, panel{(len(scroll2)-20)*4 + 1, scroll2, "", ""}, panel{5, "", "", ""})
	total := 0
	for _, p := range parts {
		total += p.duration
	}
	tick %= total
	for _, p := range parts {
		if tick < p.duration {
			label := p.label
			if len(label) > 40 {
				offset := tick / 4
				end := offset + 20
				if end > len(label) {
					end = len(label)
				}
				label = label[offset:end]
			}
			return label, p.name, p.score
		}
		tick -= p.duration
	}
	return "", "", ""
}

// INTRO fade3/fade3b change DAC entries 1,12,14 only.
func monoColor(index byte, step int) color.RGBA { return monoColorScale(index, step, 20) }
func monoColorScale(index byte, step, denominator int) color.RGBA {
	var rgb [3]int
	switch index {
	case 1:
		rgb = [3]int{52, 52, 62}
	case 12:
		rgb = [3]int{37, 37, 45}
	case 14:
		rgb = [3]int{17, 17, 25}
	}
	return color.RGBA{assets.DACRGB(byte(rgb[0] * step / denominator)), assets.DACRGB(byte(rgb[1] * step / denominator)), assets.DACRGB(byte(rgb[2] * step / denominator)), 255}
}
