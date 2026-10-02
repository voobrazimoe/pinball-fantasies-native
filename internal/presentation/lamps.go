package presentation

import "sort"

// AttractLamps replays each FLASHLIST entry's last palette packet. Sorting the
// packets preserves DOTHISFLASH's write order when two lights share DAC indices.
// It consumes the supplied counter, without advancing any gameplay light state.
func AttractLamps(tick int, list [][4]int, apply func(int, bool)) {
	type event struct {
		tick, light int
		on          bool
	}
	var events []event
	for _, f := range list {
		delay, on, off, n := f[0], f[1], f[2], f[3]
		firstOff, firstOn, period := delay+on, delay+on+off, on+off
		if period <= 0 || tick < firstOff {
			continue
		}
		e := event{firstOff, n, false}
		if tick >= firstOn {
			e.tick = firstOn + (tick-firstOn)/period*period
			e.on = true
			if e.tick+on <= tick {
				e.tick += on
				e.on = false
			}
		}
		events = append(events, e)
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].tick < events[j].tick })
	for _, e := range events {
		apply(e.light, e.on)
	}
}
