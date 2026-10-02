package hotseat

import "testing"

func TestSupportedRangeAndFactorySlots(t *testing.T) {
	for _, count := range []int{-100, 0, 1, 2, 3, 8, 9, 100} {
		var s Session[int]
		s.Initialize(count, 42)
		if s.PlayerCount < 1 || s.PlayerCount > 8 || s.CurrentPlayer != 1 {
			t.Fatal("unsafe range", count, s)
		}
		for _, v := range s.Players {
			if v != 42 {
				t.Fatal("uninitialized hidden slot")
			}
		}
		selected := s.PlayerCount
		s.Select(0)
		s.Select(9)
		if s.PlayerCount != selected {
			t.Fatal("invalid selection accepted")
		}
	}
}
