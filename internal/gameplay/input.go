// Package gameplay describes source controls independently of host devices.
package gameplay

// Controls contains held controls and one-source-tick edges. Down and Release
// name the historical keyboard plunger actions, not host keys. MouseY is a
// relative spring adjustment; its magnitude never changes the DOS task rate.
type Controls struct {
	Left, Right, Down, Release, Tilt bool
	MouseY                           int
	MouseFire                        bool
}

// Spring reproduces SPRINGSTEEN/SPRINGIT and SPRINGUP's reset. Both devices
// operate the caller's one SPRINGPOS. The table's release callback retains its
// sound and jitter semantics. Mouse input is ignored outside SPRING_VALID.
func Spring(position *uint8, valid bool, in Controls, release func(uint8)) {
	// SPRINGUP is a separate task: it uses the accumulated position and skips
	// SPRINGSTEEN/SPRINGIT on the release task, including simultaneous motion.
	if in.Release || (valid && in.MouseFire) {
		if *position > 0 {
			release(*position)
			*position = 0
		}
		return
	}
	if valid {
		if in.MouseY > 0 && *position < 32 {
			*position++
		}
		if in.MouseY < 0 && *position > 0 {
			*position--
		}
	}
	if in.Down && *position < 32 {
		*position++
	}
}
