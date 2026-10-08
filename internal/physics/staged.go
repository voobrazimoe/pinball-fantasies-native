//go:build dmoimpl1 || demodev

package physics

import "fmt"

// CandidateStage executes one explicit shared primitive without any canonical
// gameplay callbacks, camera work or Sync accounting. The candidate owns order,
// sticky failure and drain continuation. Every operation requires admission.
// sourceHold suppresses the entire step; it never writes native Ball.Hold.
func (g *Game) CandidateStage(stage string, input Inputs, sourceHold bool, gate func(string, string) error) error {
	if gate == nil {
		return fmt.Errorf("candidate consumption gate is required")
	}
	if err := gate("candidate.stage", stage); err != nil {
		return err
	}
	switch stage {
	case "scroll":
		// Fresh native reference has no screen override. Never invoke its callback.
		saved := g.ScrollForce
		g.ScrollForce = nil
		g.scroll()
		g.ScrollForce = saved

	case "early.step", "late.step":
		if !sourceHold {
			// Collision selection writes contact/pending-event fields. Inspect a
			// detached value first so an unadmitted event cannot write live state.
			probe := *g
			probe.Events = nil
			if !probe.Ball.Hold {
				probe.collision()
				for _, event := range probe.Events {
					if event.Kind != EventFlipperHit {
						return fmt.Errorf("unknown collision observation")
					}
					if err := gate("collision", "collision/flipper observation"); err != nil {
						return err
					}
				}
				if probe.pendingBumper {
					event := probe.pendingEvent
					if err := gate(fmt.Sprintf("collision kind=%d object=%d contact=(%d,%d)", event.Kind, event.Object, event.X, event.Y), "collision/OnEvent"); err != nil {
						return err
					}
					return fmt.Errorf("candidate collision event consumption is not implemented")
				}
			}
			return g.step(input)
		}
	case "early.finish":
		if g.pendingBumper {
			// Leave both pending record and public events untouched on rejection.
			if err := gate("pendingBumper", "OnEvent"); err != nil {
				return err
			}
			return fmt.Errorf("candidate event consumption is not implemented")
		}
		if g.TiltCounter > 0 {
			g.TiltCounter--
		}
		g.checkRamps()
		g.checkLevels()
	case "targets":
		if index := g.targetConsumer(); index >= 0 {
			if err := gate(fmt.Sprintf("checkSpringAndTargets target=%d", index), "target/OnEvent"); err != nil {
				return err
			}
			return fmt.Errorf("candidate target consumption is not implemented")
		}
		// The selector proved that this invocation cannot publish a callback.
		g.checkSpringAndTargets()
	default:
		return fmt.Errorf("unknown candidate physics stage %q", stage)
	}
	return nil
}

// targetConsumer is a read-only lookahead before SpringValid or HitX/Y writes.
func (g *Game) targetConsumer() int {
	b := g.Ball
	hx, hy := b.HitX, b.HitY+g.ScreenOffset
	if g.Tilted || b.High || (hx == 0 && hy == 0) {
		return -1
	}
	for i, r := range g.Table.targets {
		if r.contains(hx, hy) {
			return i
		}
	}
	return -1
}

// CandidateTargets preserves the shared selector and spring/contact/event writes.
// It admits the complete callback before those writes, with no canonical hook.
func (g *Game) CandidateTargets(preflight func(int) error, consume func(int) error) error {
	index := g.targetConsumer()
	if index >= 0 {
		if err := preflight(index); err != nil {
			return err
		}
	}
	probe := *g
	probe.OnEvent = nil
	probe.Events = nil
	probe.checkSpringAndTargets()
	g.SpringValid = probe.SpringValid
	g.Ball.HitX, g.Ball.HitY = probe.Ball.HitX, probe.Ball.HitY
	g.Events = append(g.Events, probe.Events...)
	if index >= 0 {
		return consume(index)
	}
	return nil
}
