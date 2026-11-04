package platform

type hostEvent struct {
	kind, key   int
	mouseY      int
	alt, repeat bool
}

// Suppress queued keyboard/mouse makes on focus loss, while retaining device
// identity and physical controller edges. Focus loss must precede those edges.
func focusLostEvents(events []hostEvent) []hostEvent {
	kept := make([]hostEvent, 0, len(events)+1)
	for _, event := range events {
		if event.kind == 1 {
			kept = append(kept, event)
		}
	}
	kept = append(kept, hostEvent{kind: 5})
	for _, event := range events {
		if event.kind >= 9 && event.kind <= 13 {
			kept = append(kept, event)
		}
	}
	return kept
}
