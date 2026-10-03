// Package tablelogic holds demonstrated shared FANTASIE cooperative primitives.
package tablelogic

func Add(tasks []func() bool, ids []uint64, next *uint64, f func() bool) int {
	for i := range tasks {
		if tasks[i] == nil {
			*next++
			ids[i] = *next
			tasks[i] = f
			return i
		}
	}
	panic("table task list full")
}
func Run(tasks []func() bool, ids []uint64) {
	for i := range tasks {
		if tasks[i] != nil {
			id := ids[i]
			if tasks[i]() && ids[i] == id {
				tasks[i] = nil
			}
		}
	}
}

// WAITLIST belongs to macro call sites, not individual task instances.
func Wait(counters map[string]uint16, site string, n uint16) bool {
	if counters[site] != n {
		counters[site]++
		return false
	}
	counters[site] = 0
	return true
}
func Scroll(left *uint16, phase *uint8) bool {
	for i := 0; i < 2; i++ {
		if *left == 0 {
			return true
		}
		*phase--
		if *phase == 0 {
			*phase = 8
			*left--
		}
	}
	return false
}
func Animation(header, frames []uint16, frame, loops, time *uint16) bool {
	*time--
	if *time != 0 {
		return false
	}
	old := *frame
	if old == header[2] {
		*loops--
		if *loops == 0 {
			return true
		}
		*frame = header[0]
	}
	*frame += 4
	*time = frames[old/4]
	return false
}
