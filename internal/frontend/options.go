package frontend

import (
	"image"
	"pinballfantasies/internal/assets"
)

func (v *View) optionsFrame(m *Model) *image.RGBA {
	// SHOWMENU fades the preceding page for five syncs before exposing text.
	if m.OptionTick < 5 && !m.OptionsLeaving {
		previous := *m
		previous.Mode = m.OptionsOrigin
		return v.selectorFrame(&previous)
	}
	// JULIUS slot6 installs Party's palette without the selector raster split.
	out := black(640, 240)
	p := v.Art.Tables[0].Palette
	for y := 0; y < 240; y++ {
		for x := 0; x < 128; x++ {
			q := int(v.Art.Logo.Indices[y*v.Art.Logo.Width+x]) * 3
			o := out.PixOffset(x, y)
			copy(out.Pix[o:o+3], p[q:q+3])
		}
	}
	v.sidebarInfo(out, m.SidebarTick, v.Art.OptionsInfo)
	values := m.Settings.Values()
	rows := []string{"OPTIONS MENU", "", "  BALLS:        " + values[0], "  ANGLE:        " + values[1], "  SCROLLING:    " + values[2], "  MUSIC:        " + values[3], "  RESOLUTION:   " + values[4], "", "  SAVE AND EXIT         "}
	font := v.Art.Font
	level := 63
	if m.OptionTick < 45 && !m.OptionsLeaving {
		font = v.Art.MonoFont
		level = m.OptionTick - 5
		if level < 0 {
			level = 0
		}
	}
	if m.OptionsLeaving {
		font = v.Art.MonoFont
		level = 39 - m.OptionTick
	}
	// Original fade3 uses a forty-step denominator, unlike SHOWTEXT's20.
	for i, s := range rows {
		if i >= 2 && i <= 6 {
			for len(s) < 24 {
				s += " "
			}
		}
		v.optionsText(out, font, s, 164+(24-len(s))*9, 14+i*18, level, font == v.Art.MonoFont)
	}
	if m.OptionTick >= 45 && !m.OptionsLeaving {
		y := 50 + m.OptionRow*18
		if m.OptionRow == 5 {
			y += 18
		}
		v.selectorTextFont(out, v.Art.Font, ">", 175, y, 63)
	}
	return out
}

func (v *View) optionsText(out *image.RGBA, font *assets.FrontendPicture, s string, x, y, level int, fading bool) {
	if fading {
		v.selectorTextFontFade(out, font, s, x, y, level, 40)
		return
	}
	v.selectorTextFont(out, font, s, x, y, level)
}
