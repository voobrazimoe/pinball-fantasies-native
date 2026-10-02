package platform

import "image"

// initialWindowSize is a once-only policy in host window coordinates, never
// renderer output pixels. Reserve 15% of the usable desktop for decorations and
// surrounding desktop; choose 1..4 integer gameplay scales, while accommodating
// the largest frontend presentation (640x480). User resizes subsequently win.
func initialWindowSize(usable image.Point) image.Point {
	limit := image.Pt(usable.X*17/20, usable.Y*17/20)
	scale := 1
	for candidate := 2; candidate <= 4; candidate++ {
		if candidate*320 > limit.X || candidate*383 > limit.Y {
			break
		}
		scale = candidate
	}
	size := image.Pt(max(640, 320*scale), max(480, 383*scale))
	if limit.X > 0 && size.X > limit.X {
		size.X = limit.X
	}
	if limit.Y > 0 && size.Y > limit.Y {
		size.Y = limit.Y
	}
	return size
}

// The original 640x240 selector uses doubled scanlines. Other presentations
// retain their framebuffer aspect. Backends letterbox this logical size inside the
// existing window, independently of HiDPI renderer output dimensions.
func logicalSize(frame image.Point) image.Point {
	if frame == image.Pt(640, 240) {
		return image.Pt(640, 480)
	}
	return frame
}
