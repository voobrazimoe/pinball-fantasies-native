package assets

import "image"

// SimulationFrame composes independently owned live pixels with the PF2 ball
// renderer without editing either regression snapshot.
func (p *InitialPartyLand) SimulationFrame(indices []byte, origin image.Point, viewportY int, upperForeground []byte) *image.RGBA {
	return p.SimulationFrameHeight(indices, origin, viewportY, upperForeground, 317)
}
func (p *InitialPartyLand) SimulationFrameHeight(indices []byte, origin image.Point, viewportY int, upperForeground []byte, height int) *image.RGBA {
	snapshot := *p
	field := *p.Playfield
	field.Indices = indices
	snapshot.Playfield = &field
	snapshot.State.BallOrigin = origin
	snapshot.State.Viewport = image.Rect(0, viewportY, Width, viewportY+height)
	if upperForeground != nil {
		snapshot.foreground = upperForeground
	}
	return snapshot.Framebuffer()
}
