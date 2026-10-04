package physics

import (
	"pinballfantasies/internal/settings"
	"testing"
)

func TestPresentationOverrideDoesNotChangeSourceSettings(t *testing.T) {
	for scroll := settings.ScrollHard; scroll <= settings.ScrollOff; scroll++ {
		for resolution := byte(0); resolution < 2; resolution++ {
			saved := settings.Config{ScrollMode: scroll, Resolution: resolution, Music: 1, Angle: 1}
			g := &Game{Settings: saved}
			factor := g.Settings.ScrollFactor()
			for _, portrait := range []bool{false, true, true, false, true, false} {
				g.PresentationFullTable = portrait
				render := g.PresentationSettings()
				if portrait && (render.RenderHeight() != 576 || render.MatrixY() != 0 || render.TableY() != 33) {
					t.Fatal("portrait is not matrix + full field", render)
				}
				if !portrait && render != saved {
					t.Fatal("landscape failed to restore", render, saved)
				}
				if g.Settings != saved || g.Settings.ScrollFactor() != factor {
					t.Fatal("source settings mutated")
				}
				render.ScrollMode = settings.ScrollOff // copy cannot mutate saved preference
				if g.Settings != saved {
					t.Fatal("render settings alias source settings")
				}
			}
		}
	}
}
