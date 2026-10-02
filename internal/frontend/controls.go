package frontend

import "pinballfantasies/internal/gameplay"

// Legacy fields remain for menu/replay callers; hosts submit logical Gameplay.
func (in Input) controls() gameplay.Controls {
	c := in.Gameplay
	c.Left = c.Left || in.Left
	c.Right = c.Right || in.Right
	c.Down = c.Down || in.Down
	c.Release = c.Release || in.Release
	c.Tilt = c.Tilt || in.Tilt
	return c
}
