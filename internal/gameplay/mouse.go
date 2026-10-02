package gameplay

// Mouse converts relative host motion to the DOS INT 33h Y range. INIT_MOUSE
// sets 64 mickeys per eight vertical pixels (eight counts per position) and
// bounds Y to 0..2 around MOUSEMIDDLE=1. SPRINGSTEEN consumes only the sign,
// then recentres. Excess movement cannot charge faster than one task per tick.
type Mouse struct{ remainder int }

func (m *Mouse) Motion(y int) int {
	m.remainder += y
	step := m.remainder / 8
	m.remainder %= 8
	if step > 0 {
		return 1
	}
	if step < 0 {
		return -1
	}
	return 0
}
func (m *Mouse) Clear() { m.remainder = 0 }
