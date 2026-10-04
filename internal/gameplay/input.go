// Package gameplay describes source controls independently of host devices.
package gameplay

// Controls contains held controls and one-source-tick edges. Down and Release
// name the historical keyboard plunger actions, not host keys. MouseY is a
// relative spring adjustment; its magnitude never changes the DOS task rate.
type Controls struct {
	Left, Right, Down, Release, Tilt bool
	MouseY                           int
	MouseFire                        bool
	TouchSet                         bool
	TouchTarget                      int
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
	// Additive touch input selects charge on the source task; desktop inputs
	// retain their relative adjustment and release ordering.
	if valid && in.TouchSet {
		target := in.TouchTarget
		if target < 0 {
			target = 0
		}
		if target > 32 {
			target = 32
		}
		*position = uint8(target)
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
