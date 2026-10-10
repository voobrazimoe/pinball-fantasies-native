package platform

import "testing"

func TestFocusLossPreservesControllerDiscovery(t *testing.T) {
	queued := []hostEvent{{kind: 12, key: 8}, {kind: 13, key: 1}, {kind: 9, key: 3, mouseY: 0}, {kind: 10, key: 4, mouseY: 0}, {kind: 2, key: 57}, {kind: 7}, {kind: 8}, {kind: 11}, {kind: 1}}
	got := focusLostEvents(queued)
	want := []int{5, 12, 13, 9, 10, 11, 1}
	if len(got) != len(want) {
		t.Fatal(got)
	}
	for i, kind := range want {
		if got[i].kind != kind {
			t.Fatal(got)
		}
	}
	if got[1].key != 8 || got[3].mouseY != 0 {
		t.Fatal("lost identity or release")
	}
}
